package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

// getInviteLink returns the invite link for a chat. It tries to get the invite link from the DB first. If the invite
// link is not found, it will re-generate a new link
func (bot *telegramBot) getInviteLink(chat *tb.Chat) (string, error) {
	inviteLink, err := bot.db.GetInviteLink(chat.ID)
	if err == nil {
		return inviteLink, nil
	} else if err != botdatabase.ErrInviteLinkNotFound {
		return "", err
	}

	// The link was not found in the database/cache, so we generate a new link for the chat. Note that "GetInviteLink"
	// API will actually generate a new link instead of getting the current link.
	inviteLink, err = bot.telebot.GetInviteLink(chat)
	if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
		// Chat has been migrated (why? Probably from normal groups to supergroups). We need to update some infos
		apierr, _ := err.(*tb.APIError)

		// Here we need to use Sprint() to convert migrate_to_chat_id because we don't know if it's integer or string
		// already
		newChatInfo, err := bot.telebot.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
		if err != nil {
			return "", errors.Wrap(err, "can't get chat info for migrated supergroup")
		}

		// Save the new chat info
		_ = bot.db.AddOrUpdateChat(newChatInfo)

		// Get the invite link (again! Let's hope that this is the last time...)
		inviteLink, err = bot.telebot.GetInviteLink(newChatInfo)
		if err != nil {
			return "", errors.Wrap(err, "can't get invite link from API")
		}
	} else if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == 400 || apierr.Code == 403) {
		return "", errors.Wrap(err, "no permissions for invite link")
	} else if err != nil {
		return "", errors.Wrap(err, "can't get invite link from API")
	}

	// Save the invite link in the DB/cache for later
	err = bot.db.SetInviteLink(chat.ID, inviteLink)
	if err != nil {
		bot.logger.WithError(err).WithFields(logrus.Fields{
			"chatid":     chat.ID,
			"invitelink": inviteLink,
		}).Warn("can't save invite link for chat")
	}

	return inviteLink, nil
}

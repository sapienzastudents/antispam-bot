package tbot

import (
	"fmt"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v3"
)

// getInviteLink returns the invite link for the given chat.
func (bot *telegramBot) getInviteLink(chat *tb.Chat) (string, error) {
	inviteLink, err := bot.db.GetInviteLink(chat.ID)
	if err == nil {
		return inviteLink, nil
	} else if err != botdatabase.ErrInviteLinkNotFound {
		return "", err
	}

	// The link was not found in the database/cache, so we generate a new link
	// for the chat.
	//
	// Warning: "InviteLink" API will actually generate a new link instead of
	// getting the current link.
	inviteLink, err = bot.telebot.InviteLink(chat)
	if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
		// Chat has been migrated (why? Probably from normal groups to
		// supergroups). We need to update some infos.
		apierr, _ := err.(*tb.APIError)

		ID, ok := apierr.Parameters["migrate_to_chat_id"].(int64)
		if !ok {
			return "", fmt.Errorf("migrate_to_chat_id is not an int64: %w", err)
		}
		newChatInfo, err := bot.telebot.ChatByID(ID)
		if err != nil {
			return "", fmt.Errorf("failed to get chat info for migrated supergroup: %w", err)
		}

		// Save the new chat info
		_ = bot.db.AddOrUpdateChat(newChatInfo)

		// Get the invite link (again! Let's hope that this is the last time...)
		inviteLink, err = bot.telebot.InviteLink(newChatInfo)
		if err != nil {
			return "", fmt.Errorf("failed to get invite link from API: %w", err)
		}
	} else if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == 400 || apierr.Code == 403) {
		return "", fmt.Errorf("no permissions for invite link: %w", err)
	} else if err != nil {
		return "", fmt.Errorf("failed to get invite link from API: %w", err)
	}

	// Save the invite link in the DB/cache for later
	err = bot.db.SetInviteLink(chat.ID, inviteLink)
	if err != nil {
		bot.logger.WithError(err).WithFields(logrus.Fields{
			"chatid":     chat.ID,
			"invitelink": inviteLink,
		}).Warn("Failed to save invite link for chat")
	}

	return inviteLink, nil
}

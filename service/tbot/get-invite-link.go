package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) getInviteLink(chat *tb.Chat) (string, error) {
	inviteLink, err := bot.db.GetInviteLink(chat.ID)
	if err == nil {
		return inviteLink, nil
	} else if err != botdatabase.ErrInviteLinkNotFound {
		return "", err
	}

	inviteLink, err = bot.telebot.GetInviteLink(chat)
	if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
		apierr, _ := err.(*tb.APIError)
		newChatInfo, err := bot.telebot.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
		if err != nil {
			return "", errors.Wrap(err, "can't get chat info for migrated supergroup")
		}

		_ = bot.db.UpdateMyChatroomList(newChatInfo)

		inviteLink, err = bot.telebot.GetInviteLink(newChatInfo)
		if err != nil {
			return "", errors.Wrap(err, "can't get invite link from API")
		}

		err = bot.db.SetInviteLink(chat.ID, inviteLink)
		if err != nil {
			bot.logger.WithError(err).WithFields(logrus.Fields{
				"chatid":     chat.ID,
				"invitelink": inviteLink,
			}).Warn("can't save invite link for chat")
		}
	} else if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == 400 || apierr.Code == 403) {
		return "", errors.Wrap(err, "no permissions for invite link")
	} else if err != nil {
		return "", errors.Wrap(err, "can't get invite link from API")
	}

	err = bot.db.SetInviteLink(chat.ID, inviteLink)
	if err != nil {
		bot.logger.WithError(err).WithFields(logrus.Fields{
			"chatid":     chat.ID,
			"invitelink": inviteLink,
		}).Warn("can't save invite link for chat")
	}

	return inviteLink, nil
}

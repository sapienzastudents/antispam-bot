package bot

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/database"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

// getInviteLink returns the invite link for the given chat.
func (bot *telegramBot) getInviteLink(chat *tb.Chat) (string, error) {
	inviteLink, err := bot.db.GetInviteLink(chat.ID)
	if err == nil {
		return inviteLink, nil
	} else if err != database.ErrInviteLinkNotFound {
		return "", err
	}

	// The link was not found in the database/cache, so we generate a new link
	// for the chat.
	//
	// Warning: "InviteLink" API will actually generate a new link instead of
	// getting the current link.
	inviteLink, err = bot.telebot.InviteLink(chat)
	grouperr := &tb.GroupError{}
	apierr := &tb.Error{}
	if errors.As(err, &grouperr) {
		newChatInfo, err := bot.telebot.ChatByID(grouperr.MigratedTo)
		if err != nil {
			return "", fmt.Errorf("failed to get chat info for migrated supergroup: %w", err)
		}

		// Save the new chat info
		_ = bot.db.AddChat(newChatInfo)

		// Get the invite link (again! Let's hope that this is the last time...)
		inviteLink, err = bot.telebot.InviteLink(newChatInfo)
		if err != nil {
			return "", fmt.Errorf("failed to get invite link from API: %w", err)
		}
	} else if errors.As(err, &apierr) && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
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

package tbot

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v3"
)

var (
	ErrChatNotFound = errors.New("chat not found in Telegram")
)

// DoCacheUpdate refreshes the bot cache for ALL groups, scanning one-by-one.
//
// It's a very long process: each group we scan we need to wait 2 seconds to
// avoid the Telegram rate limiter.
func (bot *telegramBot) DoCacheUpdate() error {
	startms := time.Now()
	bot.logger.Info("Chat admin scan start")

	chats, err := bot.db.ListMyChatrooms()
	if err != nil {
		return err
	}

	for _, chat := range chats {
		logfields := logrus.Fields{
			"chatid":    chat.ID,
			"chattitle": chat.Title,
		}

		if err := bot.DoCacheUpdateForChat(chat.ID); err == ErrChatNotFound {
			bot.logger.WithFields(logfields).Warning("chat not found in telegram, configuration removed")
			continue
		} else if err != nil {
			bot.logger.WithError(err).WithFields(logfields).Warning("Failed to update chat")
		}

		// Do not ask too quickly
		time.Sleep(1000 * time.Millisecond)

		members, err := bot.telebot.Len(chat)
		if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
			// We're out of the chat
			_ = bot.db.DeleteChat(chat.ID)
		} else if err != nil && strings.Contains(err.Error(), "bot is not a member of the group chat") {
			// We're out of the chat (weird errors from the library itself)
			_ = bot.db.DeleteChat(chat.ID)
		} else if err != nil {
			bot.logger.WithError(err).WithFields(logfields).Warning("Failed to get members count for chat")
		} else {
			bot.groupUserCount.WithLabelValues(strconv.FormatInt(chat.ID, 10), chat.Title).Set(float64(members))
		}

		// Do not ask too quickly
		time.Sleep(1000 * time.Millisecond)
	}

	bot.logger.Infof("Chat admin scan done in %.3f seconds", time.Since(startms).Seconds())
	return nil
}

// DoCacheUpdateForChat refreshes chat infos only for the given chat ID.
func (bot *telegramBot) DoCacheUpdateForChat(chatID int64) error {
	chat, err := bot.telebot.ChatByID(strconv.FormatInt(chatID, 10))
	if err != nil {
		if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
			_ = bot.db.DeleteChat(chatID)
			return ErrChatNotFound
		}
		return errors.Wrap(err, "failed to chat by id")
	}

	admins, err := bot.telebot.AdminsOf(chat)
	if err != nil {
		return errors.Wrap(err, "failed to get chat admins")
	}

	chatsettings, err := bot.db.GetChatSettings(chat.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get chat settings")
	}

	chatsettings.ChatAdmins.SetFromChat(admins)
	err = bot.db.SetChatSettings(chat.ID, chatsettings)
	if err != nil {
		return errors.Wrap(err, "failed to save chat settings")
	}

	return bot.db.AddOrUpdateChat(chat)
}

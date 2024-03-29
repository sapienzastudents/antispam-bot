package bot

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	tb "gopkg.in/telebot.v3"
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

	chats, err := bot.db.ListMyChats()
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
		apierr := &tb.Error{}
		if errors.As(err, &apierr) && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
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
	chat, err := bot.telebot.ChatByID(chatID)
	if err != nil {
		apierr := &tb.Error{}
		if errors.As(err, &apierr) && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
			_ = bot.db.DeleteChat(chatID)
			return ErrChatNotFound
		}
		return fmt.Errorf("failed to get chat by id: %w", err)
	}

	admins, err := bot.telebot.AdminsOf(chat)
	if err != nil {
		return fmt.Errorf("failed to get chat admins: %w", err)
	}

	chatsettings, err := bot.db.GetChatSettings(chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get chat settings: %w", err)
	}

	chatsettings.ChatAdmins.SetFromChat(admins)
	err = bot.db.SetChatSettings(chat.ID, chatsettings)
	if err != nil {
		return fmt.Errorf("failed to save chat settings: %w", err)
	}

	return bot.db.AddChat(chat)
}

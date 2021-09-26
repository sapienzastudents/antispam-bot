package tbot

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) DoCacheUpdate(g *prometheus.GaugeVec) error {
	startms := time.Now()
	bot.logger.Info("Chat admin scan start")

	chats, err := bot.db.ListMyChatrooms()
	if err != nil {
		return err
	}

	for _, chat := range chats {
		err = bot.DoCacheUpdateForChat(chat)
		if err != nil {
			bot.logger.WithError(err).WithField("chat_id", chat.ID).Warning("Error updating chat ", chat.Title)
		}

		// Do not ask too quickly
		time.Sleep(1000 * time.Millisecond)

		members, err := bot.telebot.Len(chat)
		if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
			_ = bot.db.LeftChatroom(chat.ID)
		} else if err != nil && strings.Contains(err.Error(), "bot is not a member of the group chat") {
			_ = bot.db.LeftChatroom(chat.ID)
		} else if err != nil {
			bot.logger.WithError(err).WithField("chat_id", chat.ID).Warning("Error getting members count for ", chat.Title)
		} else {
			g.WithLabelValues(fmt.Sprint(chat.ID), chat.Title).Set(float64(members))
		}

		// Do not ask too quickly
		time.Sleep(1000 * time.Millisecond)
	}

	bot.logger.Infof("Chat admin scan done in %.3f seconds", time.Since(startms).Seconds())
	return nil
}

func (bot *telegramBot) DoCacheUpdateForChat(chat *tb.Chat) error {
	newChatInfo, err := bot.telebot.ChatByID(fmt.Sprint(chat.ID))
	if err != nil {
		if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == http.StatusBadRequest || apierr.Code == http.StatusForbidden) {
			_ = bot.db.LeftChatroom(chat.ID)
			return errors.Wrap(err, fmt.Sprintf("Chat %s not found, removing configuration", chat.Title))
		}
		return errors.Wrap(err, fmt.Sprintf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error()))
	}
	chat = newChatInfo

	admins, err := bot.telebot.AdminsOf(chat)
	if err != nil {
		bot.logger.WithError(err).WithField("chat_id", chat.ID).Error("Error getting admins for chat ", chat.Title)
		return errors.Wrap(err, fmt.Sprintf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error()))
	}

	chatsettings, err := bot.db.GetChatSettings(chat.ID)
	if err != nil {
		bot.logger.WithError(err).WithField("chat_id", chat.ID).Error("Cannot get chat settings for chat ", chat.Title)
		return errors.Wrap(err, fmt.Sprintf("Cannot get chat settings for chat %d %s: %s", chat.ID, chat.Title, err.Error()))
	}

	chatsettings.ChatAdmins.SetFromChat(admins)
	err = bot.db.SetChatSettings(chat.ID, chatsettings)
	if err != nil {
		bot.logger.WithError(err).WithField("chat_id", chat.ID).Error("Cannot save chat settings for chat ", chat.Title)
		return err
	}

	return bot.db.UpdateMyChatroomList(chat)
}

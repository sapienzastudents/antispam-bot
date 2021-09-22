package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onRemoveGLine(m *tb.Message, _ botdatabase.ChatSettings) {
	if !m.Private() {
		return
	}

	payload := strings.TrimSpace(m.Text)
	if !strings.ContainsRune(payload, ' ') {
		return
	}

	parts := strings.Split(m.Text, " ")
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		_, _ = bot.telebot.Send(m.Chat, "Invalid ID specified")
		return
	}

	err = bot.db.RemoveUserBanned(userID)
	if err != nil {
		bot.logger.WithField("chatid", m.Chat.ID).WithError(err).Error("can't remove g-line")
		_, _ = bot.telebot.Send(m.Chat, "Error deleting G-Line for ID: ", err)
		return
	}
	_, _ = bot.telebot.Send(m.Chat, "OK")
}

func (bot *telegramBot) onGLine(m *tb.Message, _ botdatabase.ChatSettings) {
	_ = bot.telebot.Delete(m)
	if m.Sender.IsBot || (m.ReplyTo != nil && m.ReplyTo.Sender != nil && m.ReplyTo.Sender.IsBot) {
		return
	} else if m.ReplyTo != nil && m.ReplyTo.Sender != nil {
		if bot.db.IsGlobalAdmin(m.ReplyTo.Sender.ID) {
			bot.logger.WithField("chatid", m.Chat.ID).Warn("Won't g-line a global admin")
			return
		}
		_ = bot.telebot.Delete(m.ReplyTo)
		bot.banUser(m.Chat, m.ReplyTo.Sender)
		err := bot.db.SetUserBanned(int64(m.ReplyTo.Sender.ID))
		if err != nil {
			bot.logger.WithField("chatid", m.Chat.ID).WithError(err).Error("can't add g-line")
			return
		}

		_, _ = bot.telebot.Send(m.Sender, fmt.Sprint("GLine ok for ", m.ReplyTo.Sender))
		bot.logger.WithFields(logrus.Fields{
			"chatid":     m.Chat.ID,
			"adminid":    m.Sender.ID,
			"targetuser": m.ReplyTo.Sender.ID,
		}).Info("g-line user")
	} else if m.Text != "" {
		payload := strings.TrimSpace(m.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(m.Text, " ")
			userID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				_, _ = bot.telebot.Send(m.Chat, "Invalid ID specified")
				return
			}
			if bot.db.IsGlobalAdmin(int(userID)) {
				bot.logger.WithField("chatid", m.Chat.ID).Warn("Won't g-line a global admin")
				return
			}
			err = bot.db.SetUserBanned(userID)
			if err != nil {
				bot.logger.WithField("chatid", m.Chat.ID).WithError(err).Error("can't add g-line")
				return
			}

			_, _ = bot.telebot.Send(m.Sender, fmt.Sprint("GLine ok for ", userID))
			bot.logger.WithFields(logrus.Fields{
				"chatid":     m.Chat.ID,
				"adminid":    m.Sender.ID,
				"targetuser": userID,
			}).Info("g-line user")
		}
	}
}

func (bot *telegramBot) onSigHup(m *tb.Message, _ botdatabase.ChatSettings) {
	err := bot.db.DoCacheUpdate(bot.telebot, bot.groupUserCount)
	if err != nil {
		bot.logger.WithError(err).Warning("can't handle sighup / refresh data")
		_, _ = bot.telebot.Send(m.Chat, "Errore: "+err.Error())
	} else {
		_, _ = bot.telebot.Send(m.Chat, "Reload OK")
	}
}

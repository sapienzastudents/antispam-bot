package tbot

// G-Line (from IRC) is a global ban. When a user is g-lined, he/she is banned in any chat where the bot is. The reason
// for this command is to quickly act on trolls and spam bots.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

// onRemoveGLine acts on /remove_gline command. It removes the g-line (aka the bot ban). It does not remove the ban in
// each chat, so if the user is already banned in a chat, he/she will remain banned.
func (bot *telegramBot) onRemoveGLine(m *tb.Message, _ chatSettings) {
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

// onGLine replies to the /gline command, banning the user quoted in a group, or banning the user ID specified via a
// private message
func (bot *telegramBot) onGLine(m *tb.Message, settings chatSettings) {
	_ = bot.telebot.Delete(m)
	if m.Sender.IsBot || (m.ReplyTo != nil && m.ReplyTo.Sender != nil && m.ReplyTo.Sender.IsBot) {
		return
	} else if m.ReplyTo != nil && m.ReplyTo.Sender != nil {
		logfields := logrus.Fields{
			"chatid": m.Chat.ID,
			"userid": m.Sender.ID,
			"by":     m.ReplyTo.Sender.ID,
		}

		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.ReplyTo.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("can't check if the user is a global admin")
			return
		} else if isGlobalAdmin {
			bot.logger.WithFields(logfields).Warn("Won't g-line a global admin")
			return
		}

		bot.deleteMessage(m.ReplyTo, settings, "g-line")
		bot.banUser(m.Chat, m.ReplyTo.Sender, settings, "g-line")
		err = bot.db.SetUserBanned(int64(m.ReplyTo.Sender.ID))
		if err != nil {
			bot.logger.WithFields(logfields).WithError(err).Error("can't add g-line")
			return
		}

		_, _ = bot.telebot.Send(m.Sender, fmt.Sprint("GLine ok for ", m.ReplyTo.Sender))
		bot.logger.WithFields(logfields).Info("g-line user")
	} else if m.Text != "" && m.Private() {
		payload := strings.TrimSpace(m.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(m.Text, " ")
			userID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				_, _ = bot.telebot.Send(m.Chat, "Invalid ID specified")
				return
			}
			logfields := logrus.Fields{
				"userid": userID,
				"by":     m.Sender.ID,
			}

			isGlobalAdmin, err := bot.db.IsGlobalAdmin(int(userID))
			if err != nil {
				bot.logger.WithError(err).Error("can't check if the user is a global admin")
				return
			} else if isGlobalAdmin {
				bot.logger.WithFields(logfields).Warn("Won't g-line a global admin")
				return
			}

			err = bot.db.SetUserBanned(userID)
			if err != nil {
				bot.logger.WithFields(logfields).WithError(err).Error("can't add g-line")
				return
			}

			_, _ = bot.telebot.Send(m.Sender, fmt.Sprint("GLine ok for ", userID))
			bot.logger.WithFields(logfields).Info("g-line user")
		}
	}
}
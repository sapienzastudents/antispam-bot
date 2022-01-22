package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v3"
)

// onRemoveGLine removes the g-like (aka the bot ban). It does not remove the
// ban in each chat, so if the user is already banned in a chet, he will remain
// banned.
func (bot *telegramBot) onRemoveGLine(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}
	lang := ctx.Sender().LanguageCode

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
		_ = ctx.Send(bot.bundle.T(lang, "Invalid ID specified"))
		return
	}

	if err := bot.db.RemoveUserBanned(userID); err != nil {
		bot.logger.WithField("chatid", m.Chat.ID).WithError(err).Error("Failed to remove g-line")
		_ = ctx.Send(fmt.Sprintf(bot.bundle.T(lang, "Failed to delete G-Line for ID %d"), userID))
		return
	}
	_ = ctx.Send("OK")
}

// onGLine bans on /gline command the user quoted in a group, or bans the user
// ID given via a private message. Global admins cannot be g-lined.
//
// G-Line (from IRC) is a global ban. When a user is g-lined, he is banned in
// any chat where the bot is. The reason for this command is to quickly act on
// trolls and spam bots.
func (bot *telegramBot) onGLine(ctx tb.Context, settings chatSettings) {
	_ = ctx.Delete()

	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}
	lang := ctx.Sender().LanguageCode

	if m.Sender.IsBot || (m.ReplyTo != nil && m.ReplyTo.Sender != nil && m.ReplyTo.Sender.IsBot) {
		return
	}

	// Action on groups.
	if m.ReplyTo != nil && m.ReplyTo.Sender != nil {
		logfields := logrus.Fields{
			"chatid": m.Chat.ID,
			"userid": m.Sender.ID,
			"by":     m.ReplyTo.Sender.ID,
		}

		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.ReplyTo.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
			return
		} else if isGlobalAdmin {
			bot.logger.WithFields(logfields).Warn("Won't g-line a global admin")
			return
		}

		bot.deleteMessage(m.ReplyTo, settings, "g-line")
		bot.banUser(m.Chat, m.ReplyTo.Sender, settings, "g-line")
		if err := bot.db.SetUserBanned(m.ReplyTo.Sender.ID); err != nil {
			bot.logger.WithFields(logfields).WithError(err).Error("Failed to add g-line")
			return
		}

		_ = ctx.Send(fmt.Sprint(bot.bundle.T(lang, "G-Line ok for %d"), m.ReplyTo.Sender.ID))
		bot.logger.WithFields(logfields).Info("g-line user")
	}

	// Action on private chats.
	if m.Text != "" && m.Private() {
		payload := strings.TrimSpace(m.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(m.Text, " ")
			userID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				_ = ctx.Send(bot.bundle.T(lang, "Invalid ID specified"))
				return
			}
			logfields := logrus.Fields{"userid": userID, "by": m.Sender.ID}

			isGlobalAdmin, err := bot.db.IsGlobalAdmin(userID)
			if err != nil {
				bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
				return
			} else if isGlobalAdmin {
				bot.logger.WithFields(logfields).Warn("Won't g-line a global admin")
				return
			}

			if err := bot.db.SetUserBanned(userID); err != nil {
				bot.logger.WithFields(logfields).WithError(err).Error("can't add g-line")
				return
			}

			_ = ctx.Send(fmt.Sprint(bot.bundle.T(lang, "G-Line ok for %d"), m.ReplyTo.Sender.ID))
			bot.logger.WithFields(logfields).Info("g-line user")
		}
	}
}

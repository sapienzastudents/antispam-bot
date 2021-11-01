package tbot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v3"
)

func (bot *telegramBot) onUserLeft(ctx tb.Context, settings chatSettings) {
	msg := ctx.Message()
	if msg == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}

	bot.logger.WithFields(logrus.Fields{
		"userid":    msg.UserLeft.ID,
		"chattitle": msg.Chat.Title,
		"chatid":    msg.Chat.ID,
	}).Info("Left chat")

	// TODO: Check if !m.Private() is necessary, because this update should
	// be sent from Telegram only on groups.
	// User can be also the bot itself.
	if !msg.Private() && msg.UserLeft.ID == bot.telebot.Me.ID {
		_ = bot.db.DeleteChat(msg.Chat.ID)
	}
	if settings.OnLeaveDelete {
		if err := ctx.Delete(); err != nil {
			bot.logger.WithError(err).Error("Failed to delete leave message")
		}
	}
}

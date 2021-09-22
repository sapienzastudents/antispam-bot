package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onUserLeft(m *tb.Message, settings botdatabase.ChatSettings) {
	bot.logger.Infof("User %d left chat %s (%d)", m.UserLeft.ID, m.Chat.Title, m.Chat.ID)
	if !m.Private() && m.UserLeft.ID == bot.telebot.Me.ID {
		_ = bot.db.LeftChatroom(m.Chat.ID)
	}
	if settings.OnLeaveDelete {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot delete leave message")
		}
	}
}

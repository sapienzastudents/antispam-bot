package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onSigTerm(m *tb.Message, _ chatSettings) {
	if !m.Private() {
		_ = bot.telebot.Delete(m)
		err := bot.db.DeleteChat(m.Chat.ID)
		if err != nil {
			bot.logger.WithError(err).Error("can't delete chat info from redis")
			return
		}
		err = bot.telebot.Leave(m.Chat)
		if err != nil {
			bot.logger.WithError(err).Error("can't leave chat")
			return
		}
	}
}

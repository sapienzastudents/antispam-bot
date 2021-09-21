package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onReloadGroup(m *tb.Message, _ botdatabase.ChatSettings) {
	if !m.Private() {
		bot.botCommandsRequestsTotal.WithLabelValues("reload").Inc()

		err := bot.db.DoCacheUpdateForChat(bot.telebot, m.Chat)
		if err != nil {
			_, _ = bot.telebot.Send(m.Chat, "Error during bot reload")
			bot.logger.WithError(err).Warning("Error during bot reload")
		} else {
			_, _ = bot.telebot.Send(m.Chat, "Bot reloaded")
		}
	}
}

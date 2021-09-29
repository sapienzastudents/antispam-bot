package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// onReloadGroup is executed on /reload. It refreshes the cache for the chat where the command is issued
func (bot *telegramBot) onReloadGroup(m *tb.Message, _ chatSettings) {
	if !m.Private() {
		bot.botCommandsRequestsTotal.WithLabelValues("reload").Inc()

		err := bot.DoCacheUpdateForChat(m.Chat.ID)
		if err != nil {
			_, _ = bot.telebot.Send(m.Chat, "Error during bot reload")
			bot.logger.WithError(err).Warning("Error during bot reload")
		} else {
			_, _ = bot.telebot.Send(m.Chat, "Bot reloaded")
		}
	}
}

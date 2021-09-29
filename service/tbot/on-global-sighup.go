package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// onSigHup acts to /sighup reloading the cache for ALL chats
func (bot *telegramBot) onSigHup(m *tb.Message, _ chatSettings) {
	err := bot.DoCacheUpdate()
	if err != nil {
		bot.logger.WithError(err).Warning("can't handle sighup / refresh data")
		_, _ = bot.telebot.Send(m.Chat, "Reload error, please try later")
	} else {
		_, _ = bot.telebot.Send(m.Chat, "Reload OK")
	}
}

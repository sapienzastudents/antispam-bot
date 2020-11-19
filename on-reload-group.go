package main

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onReloadGroup(m *tb.Message, _ botdatabase.ChatSettings) {
	if !m.Private() {
		err := botdb.DoCacheUpdateForChat(b, m.Chat)
		if err != nil {
			_, _ = b.Send(m.Chat, "Error during bot reload")
			logger.WithError(err).Warning("Error during bot reload")
		} else {
			_, _ = b.Send(m.Chat, "Bot reloaded")
		}
	}
}

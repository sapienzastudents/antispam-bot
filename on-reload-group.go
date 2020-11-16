package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onReloadGroup(m *tb.Message, _ ChatSettings) {
	if !m.Private() {
		err := botdb.DoCacheUpdateForChat(m.Chat)
		if err != nil {
			b.Send(m.Chat, "Error during bot reload")
			logger.Warning("Error during bot reload: ", err.Error())
		} else {
			b.Send(m.Chat, "Bot reloaded")
		}
	}
}

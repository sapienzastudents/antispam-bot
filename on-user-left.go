package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserLeft(m *tb.Message) {
	if b, err := botdb.IsBotEnabled(m.Chat); !b || err != nil {
		return
	}
	logger.Infof("Leaving chat %s", m.Chat.Title)
	if !m.Private() && m.UserJoined.ID == b.Me.ID {
		botdb.LeftChatroom(m.Chat)
	}
}

package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserLeft(m *tb.Message) {
	logger.Infof("Leaving chat %s", m.Chat.Title)
	if !m.Private() && m.UserJoined.ID == b.Me.ID {
		redisDb.Del(fmt.Sprintf("chat.%d", m.Chat.ID))
	}
}

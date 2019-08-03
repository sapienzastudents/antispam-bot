package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUnMuteRequest(m *tb.Message) {
	if m.Private() {

		// TODO: unblock user in every chat

		_, _ = b.Send(m.Chat, "Start!")
	}
}

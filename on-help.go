package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onHelp(m *tb.Message) {
	if m.Private() {
		_, _ = b.Send(m.Chat, "Hi! If you need to unmute yourself, send /unmute to me", &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

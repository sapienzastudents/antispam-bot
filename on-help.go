package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onHelp(m *tb.Message, _ ChatSettings) {
	if m.Private() {
		_, _ = b.Send(m.Chat, "Hi!", &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

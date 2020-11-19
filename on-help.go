package main

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onHelp(m *tb.Message, _ botdatabase.ChatSettings) {
	if m.Private() {
		_, _ = b.Send(m.Chat, "Hi!", &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

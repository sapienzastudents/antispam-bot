package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

func (bot *telegramBot) setMessageExpiry(m *tb.Message, d time.Duration) {
	t := time.NewTimer(d)
	go func(m *tb.Message) {
		<-t.C
		_ = bot.telebot.Delete(m)
	}(m)
}

package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

// setMessageExpiry set an expiration for the message. In other words, the message m will be deleted after d time.
//
// This function launches a goroutine. All TTLs won't survive a bot reboot (messages won't be deleted)
func (bot *telegramBot) setMessageExpiry(m *tb.Message, d time.Duration) {
	t := time.NewTimer(d)
	go func(m *tb.Message) {
		<-t.C
		_ = bot.telebot.Delete(m)
	}(m)
}

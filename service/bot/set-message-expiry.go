package bot

import (
	"time"

	tb "gopkg.in/telebot.v3"
)

// setMessageExpiry sets the given expiration for the given message. In other
// words, the message m will be deleted after exp time.
//
// This function launches a goroutine. All TTLs won't survive a bot reboot
// (messages won't be deleted).
func (bot *telegramBot) setMessageExpiry(m *tb.Message, exp time.Duration) {
	t := time.NewTimer(exp)
	go func(m *tb.Message) {
		<-t.C
		_ = bot.telebot.Delete(m)
	}(m)
}

package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

// checkGroupAdmin is a "firewall" for group admin only functions
func (bot *telegramBot) checkGroupAdmin(actionHandler func(*tb.Message, botdatabase.ChatSettings)) func(*tb.Message, botdatabase.ChatSettings) {
	return func(m *tb.Message, settings botdatabase.ChatSettings) {
		if m.Private() || (!m.Private() && settings.ChatAdmins.IsAdmin(m.Sender)) || bot.db.IsGlobalAdmin(m.Sender.ID) {
			actionHandler(m, settings)
			return
		}
		_ = bot.telebot.Delete(m)
		msg, _ := bot.telebot.Send(m.Chat, "Sorry, only group admins can use this command")
		bot.setMessageExpiry(msg, 10*time.Second)
	}
}

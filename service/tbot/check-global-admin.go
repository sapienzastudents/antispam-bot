package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// checkGlobalAdmin is a "firewall" for global admin only functions
func (bot *telegramBot) checkGlobalAdmin(actionHandler func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !bot.db.IsGlobalAdmin(m.Sender.ID) {
			return
		}
		actionHandler(m)
	}
}

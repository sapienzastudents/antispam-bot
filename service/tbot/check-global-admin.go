package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// checkGlobalAdmin is a "firewall" wrapper for global admin only handlers. It checks if the sender is a global admin
func (bot *telegramBot) checkGlobalAdmin(actionHandler func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("can't check if the user is a global admin")
			return
		} else if !isGlobalAdmin {
			return
		}
		actionHandler(m)
	}
}

// globalAdminHandler register a new handler that is available only to global admins
func (bot *telegramBot) globalAdminHandler(endpoint interface{}, fn contextualChatSettingsFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.checkGlobalAdmin(bot.refreshDBInfo(fn))))
}

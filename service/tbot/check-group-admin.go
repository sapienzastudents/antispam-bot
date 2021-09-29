package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

// checkGroupAdmin is a "firewall" wrapper for group admin only handlers. It checks if the sender is an admin of the
// chat, or a global admin
func (bot *telegramBot) checkGroupAdmin(actionHandler func(*tb.Message, chatSettings)) func(*tb.Message, chatSettings) {
	return func(m *tb.Message, settings chatSettings) {
		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("can't check if the user is a global admin")
			return
		}

		if m.Private() || (!m.Private() && settings.ChatAdmins.IsAdmin(m.Sender)) || isGlobalAdmin {
			actionHandler(m, settings)
			return
		}
		_ = bot.telebot.Delete(m)
		msg, _ := bot.telebot.Send(m.Chat, "Sorry, only group admins can use this command")
		bot.setMessageExpiry(msg, 10*time.Second)
	}
}

// chatAdminHandler register a new handler that is available only to groups admins or global admins
func (bot *telegramBot) chatAdminHandler(endpoint interface{}, fn contextualChatSettingsFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.refreshDBInfo(bot.checkGroupAdmin(fn))))
}

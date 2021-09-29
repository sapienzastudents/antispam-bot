package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// checkGlobalAdmin is a "firewall" for global admin only functions
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

func (bot *telegramBot) globalAdminHandler(endpoint interface{}, fn refreshDBInfoFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.checkGlobalAdmin(bot.refreshDBInfo(fn))))
}

package tbot

import tb "gopkg.in/tucnak/telebot.v3"

// checkGlobalAdmin is a "firewall" wrapper for global admin only handlers. It
// checks if the sender is a global admin.
func (bot *telegramBot) checkGlobalAdmin(actionHandler tb.HandlerFunc) tb.HandlerFunc {
	return func(ctx tb.Context) error {
		m := ctx.Message()
		if m == nil {
			bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
			return nil
		}

		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
			return nil
		} else if !isGlobalAdmin {
			return nil
		}
		return actionHandler(ctx)
	}
}

// globalAdminHandler register a new handler that is available only to global
// admins.
func (bot *telegramBot) globalAdminHandler(endpoint interface{}, fn contextualChatSettingsFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.checkGlobalAdmin(bot.refreshDBInfo(fn))))
}

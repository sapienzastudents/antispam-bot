package bot

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v3"
)

// checkGroupAdmin returns a function suited to be passed to refreshDBInfo.
//
// It is a "firewall" wrapper for group admin only handlers. It checks if the
// sender is an admin of the chat or a global admin.
func (bot *telegramBot) checkGroupAdmin(actionHandler contextualChatSettingsFunc) contextualChatSettingsFunc {
	return func(ctx tb.Context, settings chatSettings) {
		m := ctx.Message()
		if m == nil {
			bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
			return
		}

		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
			return
		}

		if m.Private() || (!m.Private() && settings.ChatAdmins.IsAdmin(m.Sender)) || isGlobalAdmin {
			actionHandler(ctx, settings)
			return
		}
		_ = ctx.Delete()
		lang := ctx.Sender().LanguageCode
		msg, _ := bot.telebot.Send(m.Chat, bot.bundle.T(lang, "Sorry, only group admins can use this command"))
		bot.setMessageExpiry(msg, 10*time.Second)
	}
}

// chatAdminHandler registers a new handler that is available only to groups
// admins or global admins.
func (bot *telegramBot) chatAdminHandler(endpoint interface{}, fn contextualChatSettingsFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.refreshDBInfo(bot.checkGroupAdmin(fn))))
}

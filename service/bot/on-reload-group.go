package bot

import tb "gopkg.in/tucnak/telebot.v3"

// onReloadGroup refreshes the cache for the group where /reload command is sent.
func (bot *telegramBot) onReloadGroup(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}
	lang := ctx.Sender().LanguageCode

	// /reload is useless on private chat...
	if !m.Private() {
		bot.botCommandsRequestsTotal.WithLabelValues("reload").Inc()

		err := bot.DoCacheUpdateForChat(m.Chat.ID)
		if err != nil {
			_ = ctx.Send(bot.bundle.T(lang, "An error has been detected during reload, contact an administrator!"))
			bot.logger.WithError(err).Warning("Failed to refresh cache")
		} else {
			_ = ctx.Send(bot.bundle.T(lang, "Bot reloaded!"))
		}
	}
}

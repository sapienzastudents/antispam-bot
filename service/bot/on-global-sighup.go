package bot

import tb "gopkg.in/telebot.v3"

// onSigHup refreshes the cache for ALL groups on /sighup command.
func (bot *telegramBot) onSigHup(ctx tb.Context, settings chatSettings) {
	lang := ctx.Sender().LanguageCode

	if err := bot.DoCacheUpdate(); err != nil {
		bot.logger.WithError(err).Warning("Failed to handle sighup / refresh data")
		_ = ctx.Send(bot.bundle.T(lang, "Reload error, please try later"))
	} else {
		_ = ctx.Send(bot.bundle.T(lang, "Reload OK"))
	}
}

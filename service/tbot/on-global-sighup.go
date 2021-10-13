package tbot

import tb "gopkg.in/tucnak/telebot.v3"

// onSigHup refreshes the cache for ALL groups on /sighup command.
func (bot *telegramBot) onSigHup(ctx tb.Context, settings chatSettings) {
	if err := bot.DoCacheUpdate(); err != nil {
		bot.logger.WithError(err).Warning("Failed to handle sighup / refresh data")
		_ = ctx.Send("Reload error, please try later")
	} else {
		_ = ctx.Send("Reload OK")
	}
}

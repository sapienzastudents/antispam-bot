package bot

import tb "gopkg.in/telebot.v3"

// onDont is fired on /dont command. It works only if the command is given as a
// reply for another message, it deletes the command message and replies to the
// original message.
func (bot *telegramBot) onDont(ctx tb.Context, settings chatSettings) {
	msg := ctx.Message()
	if msg == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}

	defer func() {
		if err := ctx.Delete(); err != nil {
			bot.logger.WithError(err).Error("Failed to delete message")
		}
	}()

	if !msg.IsReply() {
		return
	}

	lang := ctx.Sender().LanguageCode
	m := "https://dontasktoask.com\n" + bot.bundle.T(lang, "Don't ask to ask, just ask!")
	if _, err := bot.telebot.Reply(msg.ReplyTo, m); err != nil {
		bot.logger.WithError(err).Error("Failed to reply")
		return
	}
}

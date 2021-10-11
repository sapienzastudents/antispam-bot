package tbot

import tb "gopkg.in/tucnak/telebot.v3"

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

	_, err := bot.telebot.Reply(msg.ReplyTo, "https://dontasktoask.com\nNon chiedere di chiedere, chiedi pure :)")
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply")
		return
	}
}

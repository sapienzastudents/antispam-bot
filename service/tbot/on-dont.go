package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// onDont fires when /dont command is used
func (bot *telegramBot) onDont(m *tb.Message, _ chatSettings) {
	defer func() {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to delete message")
		}
	}()

	if !m.IsReply() {
		return
	}

	_, err := bot.telebot.Reply(m.ReplyTo, "https://dontasktoask.com\nNon chiedere di chiedere, chiedi pure :)")
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply")
		return
	}
}

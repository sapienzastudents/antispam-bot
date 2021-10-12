package tbot

import tb "gopkg.in/tucnak/telebot.v3"

// onSigTerm quits from the group where the command /sigterm is sent and deletes
// all infos about it.
func (bot *telegramBot) onSigTerm(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}

	// /sigterm is useless on private chats...
	if !m.Private() {
		_ = ctx.Delete()
		if err := bot.db.DeleteChat(m.Chat.ID); err != nil {
			bot.logger.WithError(err).Error("Failed to delete chat info from redis")
			return
		}
		if err := bot.telebot.Leave(m.Chat); err != nil {
			bot.logger.WithError(err).Error("Failed to leave chat")
			return
		}
	}
}

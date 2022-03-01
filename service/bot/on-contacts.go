package bot

import tb "gopkg.in/telebot.v3"

// onContacts sends a small message with link to website and repository.
func (bot *telegramBot) onContacts(ctx tb.Context) error {
	sender := ctx.Sender()
	msgToEdit := ctx.Message()

	lang := sender.LanguageCode

	// Message with info.
	msg := bot.bundle.T(lang, "<b>Contacts</b>\n\n") +
		bot.bundle.T(lang, "You can reach us on our <a href=\"https://sapienzahub.it/\">SapienzaHub website</a> for more information.\n\n") +
		bot.bundle.T(lang, "If you have any problem with the bot, open an issue on the our <a href=\"https://gitlab.com/sapienzastudents/antispam-telegram-bot/\">GitLab repository</a>!")

	// Close button.
	bt := tb.InlineButton{
		Unique: "on-contacts-close",
		Text:   "🚪 " + bot.bundle.T(lang, "Close"),
	}
	var chatButtons [][]tb.InlineButton
	chatButtons = append(chatButtons, []tb.InlineButton{bt})
	bot.telebot.Handle(&bt, func(ctx tb.Context) error {
		callback := ctx.Callback()
		_ = bot.telebot.Respond(callback)
		_ = bot.telebot.Delete(callback.Message)
		return nil
	})

	options := &tb.SendOptions{
		ParseMode:   tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{InlineKeyboard: chatButtons},
	}
	_, err := bot.telebot.Edit(msgToEdit, msg, options)
	return err
}

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

	// "Back" button.
	backBtn := tb.InlineButton{
		Unique: "on_contacts_back",
		Text:   "◀ " + bot.bundle.T(lang, "Back"),
	}
	bot.telebot.Handle(&backBtn, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}
		callback := ctx.Callback()
		bot.sendHelpMessage(callback.Sender, callback.Message)
		return nil
	})
	var chatButtons [][]tb.InlineButton
	chatButtons = append(chatButtons, []tb.InlineButton{backBtn})

	options := &tb.SendOptions{
		ParseMode:   tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{InlineKeyboard: chatButtons},
	}
	_, err := bot.telebot.Edit(msgToEdit, msg, options)
	return err
}

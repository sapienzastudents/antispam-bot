package bot

import tb "gopkg.in/telebot.v3"

// onGuide fires when guide button is pressed.
func (bot *telegramBot) onGuide(ctx tb.Context) {
	m := ctx.Message()
	// This action is fired on button pressed, so we change "page" in the
	// message. Uses can go back pressing "Close" button.
	defer func() {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to delete message")
		}
	}()

	lang := ctx.Sender().LanguageCode

	// "Back" button.
	bt := tb.InlineButton{
		Unique: "on_guide_back",
		Text:   "◀ " + bot.bundle.T(lang, "Back"),
	}
	bot.telebot.Handle(&bt, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}
		callback := ctx.Callback()
		bot.sendHelpMessage(callback.Sender, callback.Message)
		return nil
	})
	chatButtons := [][]tb.InlineButton{{bt}}

	msg := bot.bundle.T(lang, "What you need to do to add a group on the network:\n\n") +
		bot.bundle.T(lang, "<b>0.</b> Check if your group is already on the list;\n") +
		bot.bundle.T(lang, "<b>1.</b> Create the group;\n") +
		bot.bundle.T(lang, "<b>2.</b> Add this bot as admin with all permissions <b>except</b> for anonymous;\n") +
		bot.bundle.T(lang, "<b>3.</b> Write to the bot with <code>/start</code> command, go to <code>Settings</code>, select the chat you've just added, then click on <code>Modify category</code> (✏️) and follow the instructions on the message.\n\n") +
		bot.bundle.T(lang, "Thank you for joining the community!") + " 🙏"

	sendOptions := &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: chatButtons,
		},
	}

	_, err := bot.telebot.Send(m.Chat, msg, sendOptions)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply")
	}
}

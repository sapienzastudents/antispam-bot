package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// onGuide fires when guide button is pressed.
func (bot *telegramBot) onGuide(m *tb.Message) {
	// This action is fired on button pressed, so we change "page" in the
	// message. Uses can go back pressing "Close" button.
	defer func() {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to delete message")
		}
	}()

	bt := tb.InlineButton{
		Unique: "on-guide-close",
		Text:   "✖️ Close / Chiudi",
	}

	var chatButtons [][]tb.InlineButton
	chatButtons = append(chatButtons, []tb.InlineButton{bt})
	bot.telebot.Handle(&bt, func(callback *tb.Callback) {
		_ = bot.telebot.Respond(callback)
		_ = bot.telebot.Delete(callback.Message)
	})

	const message = `
	Ecco cosa devi fare per aggiungere il gruppo alla rete:

	<b>0.</b> Controlla se il gruppo è già presente nell'elenco;
	<b>1.</b> Crea il gruppo;
	<b>2.</b> Aggiungi il bot come amministratore con tutti permessi <b>tranne</b> quello di rimanere anonimo;
	<b>3.</b> Scrivi al bot inviando il comando <code>/start</code>, vai su <code>Impostazioni</code>, seleziona la chat appena creata, quindi clicca su <code>Modifica categoria(✏️)</code> e segui le istruzioni indicate nel messaggio.

	Grazie per esserti unito alla comunità!
	`

	sendOptions := &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: chatButtons,
		},
	}

	_, err := bot.telebot.Send(m.Chat, message, sendOptions)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply")
	}
}

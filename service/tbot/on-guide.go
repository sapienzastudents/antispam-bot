package tbot

import tb "gopkg.in/tucnak/telebot.v2"

// onGuide fires when guide button is pressed
func (bot *telegramBot) onGuide(m *tb.Message) {
	// cancella il messaggio originale dell'utente
	defer func() {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to delete message")
		}
	}()

	// == BOTTONE CHIUDI ==
	var bt = tb.InlineButton{
		Unique: "groups_settings_list_close",
		Text:   "✖️ Close / Chiudi",
	}

	var chatButtons [][]tb.InlineButton

	chatButtons = append(chatButtons, []tb.InlineButton{bt})
	bot.telebot.Handle(&bt, func(callback *tb.Callback) {
		_ = bot.telebot.Respond(callback)
		_ = bot.telebot.Delete(callback.Message)
	})

	var sendOptions = tb.SendOptions{}
	sendOptions = tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: chatButtons,
		},
	}

	// invia l'effettivo messaggio
	_, err := bot.telebot.Send(m.Chat, "Ecco cosa devi fare per aggiungere il gruppo al network:\n\n0. Controlla se il gruppo è già presente\n1. Crea il gruppo\n2. Aggiungi il bot come admin, dandogli tutti i permessi TRANNE rimanere anonimo\n3. Scrivi al bot, digita /start > Impostazioni > la chat appena creata > \"Modifica categoria (✏️)\" e segui le istruzioni riportate nel messaggio lì.\n\nGrazie per esserti unito al network!🙏", &sendOptions)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply")
		return
	}
}

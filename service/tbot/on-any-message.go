package tbot

import (
	"strings"
	"unicode/utf8"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onAnyMessage(m *tb.Message, settings chatSettings) {
	// Check the user is entering free text for changing category
	state := bot.getStateFor(m.Sender, m.Chat)
	if state.AddGlobalCategory || state.AddSubCategory {
		// Load current chat settings
		settings, err := bot.getChatSettings(state.ChatToEdit)
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Warn("can't get chat settings")
			return
		}

		// Change category/subcategory
		if state.AddGlobalCategory {
			categories := strings.Split(m.Text, "\n")
			settings.MainCategory = categories[0]
			if len(categories) > 1 {
				settings.SubCategory = categories[1]
			} else {
				settings.SubCategory = ""
			}
		} else {
			categories := strings.Split(m.Text, "\n")
			settings.SubCategory = categories[0]
		}

		// Max 64 is because Telegram accepts 64 chars MAX as callback argument
		if utf8.RuneCountInString(settings.MainCategory) > 64 || utf8.RuneCountInString(settings.SubCategory) > 64 {
			_, _ = bot.telebot.Reply(m, "Nome troppo lungo, massimo 64 caratteri")
			return
		}

		// Save chat settings
		err = bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Warn("can't save chat settings")
		}

		// Reset category naming flags in user state
		state.AddSubCategory = false
		state.AddGlobalCategory = false
		state.Save()

		// Button for opening the settings menu again
		settingsbt := tb.InlineButton{
			Text:   "Torna alle impostazioni",
			Unique: "back_to_settings",
		}
		bot.handleAdminCallbackStateful(&settingsbt, bot.backToSettingsFromCallback)

		_, _ = bot.telebot.Send(m.Chat, "Categoria salvata", &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{settingsbt},
			},
		})

		return
	}

	// Check the message against the antispam system
	if !m.Private() {
		if banned, err := bot.db.IsUserBanned(int64(m.Sender.ID)); err == nil && banned {
			bot.banUser(m.Chat, m.Sender, settings, "user g-lined")
			bot.deleteMessage(m, settings, "user g-lined")
			return
		}

		if bot.cas.IsBanned(m.Sender.ID) {
			bot.performAction(m, m.Sender, settings, settings.OnBlacklistCAS, "CAS banned")
			return
		}

		textvalues := []string{
			m.Text,
			m.Caption,
			m.Payload,
		}
		if m.Photo != nil {
			textvalues = append(textvalues, m.Photo.Caption)
		}
		if m.Audio != nil {
			textvalues = append(textvalues, m.Audio.Caption)
		}
		if m.Document != nil {
			textvalues = append(textvalues, m.Document.Caption)
		}
		if m.Video != nil {
			textvalues = append(textvalues, m.Video.Caption)
		}

		bot.spamFilter(m, settings, textvalues)
	}
}

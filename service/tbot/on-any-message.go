package tbot

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onAnyMessage(m *tb.Message, settings chatSettings) {
	if m.Private() && m.OriginalSender != nil && bot.db.IsGlobalAdmin(m.Sender.ID) {
		_, _ = bot.telebot.Send(m.Chat, fmt.Sprint(m.OriginalSender))
		return
	}

	chatEditVal, isForEdit := bot.globaleditcat.Get(fmt.Sprint(m.Sender.ID))
	if isForEdit {
		chatEdit, chatEditOk := chatEditVal.(inlineCategoryEdit)
		if chatEditOk && m.Text != "" && (m.Private() || m.Chat.ID == chatEdit.ChatID) {
			bot.globaleditcat.Delete(fmt.Sprint(m.Sender.ID))
			chat, err := bot.telebot.ChatByID(fmt.Sprint(chatEdit.ChatID))
			if err != nil {
				bot.logger.WithError(err).WithField("chat", chatEdit.ChatID).Warn("can't get chat info")
				return
			}

			settings, err := bot.db.GetChatSetting(bot.telebot, chat)
			if err != nil {
				bot.logger.WithError(err).WithField("chat", chat.ID).Warn("can't get chat settings")
				return
			}

			if chatEdit.Category == "" {
				categories := strings.Split(m.Text, "\n")
				settings.MainCategory = categories[0]
				if len(categories) > 1 {
					settings.SubCategory = categories[1]
				} else {
					settings.SubCategory = ""
				}
			} else {
				categories := strings.Split(m.Text, "\n")
				settings.MainCategory = chatEdit.Category
				settings.SubCategory = categories[0]
			}

			err = bot.db.SetChatSettings(chat.ID, settings)
			if err != nil {
				bot.logger.WithError(err).WithField("chat", chat.ID).Warn("can't save chat settings")
			}

			settingsbt := tb.InlineButton{
				Text:   "Torna alle impostazioni",
				Unique: "back_to_settings",
				Data:   fmt.Sprint(chat.ID),
			}

			bot.telebot.Handle(&settingsbt, bot.backToSettingsFromCallback)

			_, _ = bot.telebot.Send(m.Chat, "Categoria salvata", &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{
					{settingsbt},
				},
			})

			return
		}
	}

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

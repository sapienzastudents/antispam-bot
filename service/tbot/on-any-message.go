package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/antispam"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onAnyMessage(m *tb.Message, settings botdatabase.ChatSettings) {
	if m.Private() && m.OriginalSender != nil && bot.db.IsGlobalAdmin(m.Sender) {
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

			err = bot.db.SetChatSettings(chat, settings)
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

	// Note: this will not scale very well - keep an eye on it
	if !m.Private() {
		if banned, err := bot.db.IsUserBanned(int64(m.Sender.ID)); err == nil && banned {
			bot.logger.Infof("User %d banned, performing ban + message deletion", m.Sender.ID)
			_ = bot.telebot.Delete(m)
			bot.banUser(m.Chat, m.Sender)
			return
		}

		if settings.OnBlacklistCAS.Action != botdatabase.ACTION_NONE && bot.cas.IsBanned(m.Sender.ID) {
			bot.logger.Infof("User %d CAS-banned, performing action: %s", m.Sender.ID, prettyActionName(settings.OnBlacklistCAS))
			_ = bot.telebot.Delete(m)
			bot.performAction(m, m.Sender, settings.OnBlacklistCAS)
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

		for _, text := range textvalues {
			if settings.OnMessageChinese.Action != botdatabase.ACTION_NONE {
				chinesePercent := antispam.ChineseChars(text)
				bot.logger.Debugf("SPAM detection (msg id %d): chinese %f", m.ID, chinesePercent)
				if chinesePercent > 0.05 {
					_ = bot.telebot.Delete(m)
					bot.performAction(m, m.Sender, settings.OnMessageChinese)
					return
				}
			}

			if settings.OnMessageArabic.Action != botdatabase.ACTION_NONE {
				arabicPercent := antispam.ArabicChars(text)
				bot.logger.Debugf("SPAM detection (msg id %d): arabic %f", m.ID, arabicPercent)
				if arabicPercent > 0.05 {
					_ = bot.telebot.Delete(m)
					bot.performAction(m, m.Sender, settings.OnMessageArabic)
					return
				}
			}
		}
	}
}

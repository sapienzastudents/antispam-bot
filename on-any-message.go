package main

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onAnyMessage(m *tb.Message, settings botdatabase.ChatSettings) {
	// Note: this will not scale very well - keep an eye on it
	if !m.Private() {
		if settings.OnBlacklistCAS.Action != botdatabase.ACTION_NONE && IsCASBanned(m.Sender.ID) {
			logger.Infof("User %d CAS-banned, performing action: %s", m.Sender.ID, prettyActionName(settings.OnBlacklistCAS))
			performAction(m, m.Sender, settings.OnBlacklistCAS)
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
				chinesePercent := chineseChars(text)
				logger.Debugf("SPAM detection (msg id %d): chinese %f", m.ID, chinesePercent)
				if chinesePercent > 0.05 {
					performAction(m, m.Sender, settings.OnMessageChinese)
					return
				}
			}

			if settings.OnMessageArabic.Action != botdatabase.ACTION_NONE {
				arabicPercent := arabicChars(text)
				logger.Debugf("SPAM detection (msg id %d): arabic %f", m.ID, arabicPercent)
				if arabicPercent > 0.05 {
					performAction(m, m.Sender, settings.OnMessageArabic)
					return
				}
			}
		}
	}
}

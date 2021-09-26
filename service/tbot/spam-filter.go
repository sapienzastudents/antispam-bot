package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/antispam"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) spamFilter(m *tb.Message, settings chatSettings, textvalues []string) {
	for _, text := range textvalues {
		if settings.OnMessageChinese.Action != botdatabase.ActionNone {
			chinesePercent := antispam.ChineseChars(text)
			bot.logger.Debugf("SPAM detection (msg id %d): chinese %f", m.ID, chinesePercent)
			if chinesePercent > 0.05 {
				bot.performAction(m, m.Sender, settings, settings.OnMessageChinese, "Chinese message filter enabled")
				return
			}
		}

		if settings.OnMessageArabic.Action != botdatabase.ActionNone {
			arabicPercent := antispam.ArabicChars(text)
			bot.logger.Debugf("SPAM detection (msg id %d): arabic %f", m.ID, arabicPercent)
			if arabicPercent > 0.05 {
				bot.performAction(m, m.Sender, settings, settings.OnMessageChinese, "Arabic message filter enabled")
				return
			}
		}
	}
}

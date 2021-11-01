package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/antispam"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v3"
)

// spamFilter checks all text values in the slice against antispam functions. If
// a violation is detected, the corresponding action will be performed.
//
// Example: if the action on Chinese messages is delete, the bot will delete the
// message.
//
// Time complexity: O(n*2m) where n is the length of the textvalues slice, m is
// the length of the longest string in the slice
func (bot *telegramBot) spamFilter(m *tb.Message, settings chatSettings, textvalues []string) {
	for _, text := range textvalues {
		// Note: nothing personal. We were forced to write these blocks for
		// chinese texts in a period of time when bots were targetting our
		// group. This check is trying to avoid banning people randomly just for
		// having chinese/arabic names, however false positive might arise.
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

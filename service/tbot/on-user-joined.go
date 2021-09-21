package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/antispam"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onUserJoined(m *tb.Message, settings botdatabase.ChatSettings) {
	if m.IsService() && !m.Private() && m.UserJoined.ID == bot.telebot.Me.ID {
		bot.logger.Infof("Joining chat %s", m.Chat.Title)
		return
	}

	bot.logger.Debugf("User %d (%s %s %s) joined chat %s (%d)", m.UserJoined.ID, m.UserJoined.Username,
		m.UserJoined.FirstName, m.UserJoined.LastName, m.Chat.Title, m.Chat.ID)

	if settings.OnBlacklistCAS.Action != botdatabase.ACTION_NONE && settings.OnBlacklistCAS.Action != botdatabase.ACTION_DELETE_MSG && bot.cas.IsBanned(m.Sender.ID) {
		bot.logger.Infof("User %d CAS-banned, performing action: %s", m.Sender.ID, prettyActionName(settings.OnBlacklistCAS))
		bot.performAction(m, m.Sender, settings.OnBlacklistCAS)
		return
	}

	// Note: nothing personal. We were forced to write these blocks for chinese texts in a period of time when bots were
	// targetting our group. This check is trying to avoid banning people randomly just for having chinese/arabic names,
	// however false positive might arise
	textvalues := []string{
		m.UserJoined.Username,
		m.UserJoined.FirstName,
		m.UserJoined.LastName,
	}

	for _, text := range textvalues {
		if settings.OnJoinChinese.Action != botdatabase.ACTION_NONE {
			chinesePercent := antispam.ChineseChars(text)
			bot.logger.Debugf("SPAM detection (%s): chinese %f", text, chinesePercent)
			if chinesePercent > 0.5 {
				bot.performAction(m, m.UserJoined, settings.OnJoinChinese)
				return
			}
		}

		if settings.OnJoinArabic.Action != botdatabase.ACTION_NONE {
			arabicPercent := antispam.ArabicChars(text)
			bot.logger.Debugf("SPAM detection (%s): arabic %f", text, arabicPercent)
			if arabicPercent > 0.5 {
				bot.performAction(m, m.UserJoined, settings.OnJoinArabic)
				return
			}
		}
	}

	if settings.OnJoinDelete {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot delete join message")
		}
	}
}

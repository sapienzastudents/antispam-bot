package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserJoined(m *tb.Message, settings ChatSettings) {
	if m.IsService() && !m.Private() && m.UserJoined.ID == b.Me.ID {
		logger.Infof("Joining chat %s", m.Chat.Title)
		return
	}

	logger.Debugf("User %d (%s %s %s) joined chat %s (%d)", m.UserJoined.ID, m.UserJoined.Username,
		m.UserJoined.FirstName, m.UserJoined.LastName, m.Chat.Title, m.Chat.ID)

	textvalues := []string{
		m.UserJoined.Username,
		m.UserJoined.FirstName,
		m.UserJoined.LastName,
	}

	for _, text := range textvalues {
		if settings.OnJoinChinese.Action != ACTION_NONE {
			chinesePercent := chineseChars(text)
			logger.Debugf("SPAM detection (%s): chinese %f", text, chinesePercent)
			if chinesePercent > 0.5 {
				performAction(m, m.UserJoined, settings.OnJoinChinese)
				return
			}
		}

		if settings.OnJoinArabic.Action != ACTION_NONE {
			arabicPercent := arabicChars(text)
			logger.Debugf("SPAM detection (%s): arabic %f", text, arabicPercent)
			if arabicPercent > 0.5 {
				performAction(m, m.UserJoined, settings.OnJoinArabic)
				return
			}
		}
	}

	if settings.OnJoinDelete {
		err := b.Delete(m)
		if err != nil {
			logger.Critical("Cannot delete join message: ", err)
		}
	}
}

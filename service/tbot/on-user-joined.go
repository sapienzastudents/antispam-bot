package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onUserJoined(m *tb.Message, settings chatSettings) {
	if m.IsService() && !m.Private() && m.UserJoined.ID == bot.telebot.Me.ID {
		bot.logger.WithField("chatid", m.Chat.ID).Info("Joining chat")
		return
	}

	if banned, err := bot.db.IsUserBanned(int64(m.Sender.ID)); err == nil && banned {
		bot.banUser(m.Chat, m.Sender, settings, "user g-lined")
		bot.deleteMessage(m, settings, "user g-lined")
		return
	}

	if bot.cas.IsBanned(m.Sender.ID) {
		bot.casDatabaseMatch.Inc()
		bot.performAction(m, m.Sender, settings, settings.OnBlacklistCAS, "CAS banned")
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
	bot.spamFilter(m, settings, textvalues)

	if settings.OnJoinDelete {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot delete join message")
		}
	}
}

package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// onAddedToGroup is fired when the bot is added to a group
func (bot *telegramBot) onAddedToGroup(m *tb.Message, _ chatSettings) {
	bot.logger.WithField("chatid", m.Chat.ID).Info("Joining chat")
	// Do nothing: the previous chained handler (refreshDBInfo) will take care of creating the new chat in the DB
}

// onUserJoined is fired when a user is added (or joins) to a group
func (bot *telegramBot) onUserJoined(m *tb.Message, settings chatSettings) {
	// If it's me (the bot) joining the chat, don't do anything else
	// Note: this should be replaced with onAddedToGroup
	// TODO: verify if this code is required (might be replaced by onAddedToGroup)
	if m.IsService() && !m.Private() && m.UserJoined.ID == bot.telebot.Me.ID {
		bot.logger.WithField("chatid", m.Chat.ID).Info("Joining chat")
		return
	}

	// Check if the user that's joining is g-lined. If so, ban them and delete the join service message
	if banned, err := bot.db.IsUserBanned(int64(m.Sender.ID)); err == nil && banned {
		bot.banUser(m.Chat, m.Sender, settings, "user g-lined")
		bot.deleteMessage(m, settings, "user g-lined")
		return
	}

	// Check if the user that's joining is CAS banned. If so, do the proper action
	if bot.cas != nil && bot.cas.IsBanned(m.Sender.ID) {
		bot.casDatabaseMatch.Inc()
		bot.performAction(m, m.Sender, settings, settings.OnBlacklistCAS, "CAS banned")
		return
	}

	// Check for spam items in user names
	textvalues := []string{
		m.UserJoined.Username,
		m.UserJoined.FirstName,
		m.UserJoined.LastName,
	}
	bot.spamFilter(m, settings, textvalues)

	// If the owner wants to delete all join messages, do so
	if settings.OnJoinDelete {
		err := bot.telebot.Delete(m)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot delete join message")
		}
	}
}

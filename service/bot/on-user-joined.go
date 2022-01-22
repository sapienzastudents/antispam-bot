package bot

import tb "gopkg.in/tucnak/telebot.v3"

// onAddedToGroup is fired when the bot is added to a group.
func (bot *telegramBot) onAddedToGroup(ctx tb.Context, settings chatSettings) {
	chat := ctx.Chat()
	if chat == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Chat, ignored")
		return
	}
	// Do nothing: the previous chained handler (refreshDBInfo) will take care
	// of creating the new chat in the DB.
	bot.logger.WithField("chatid", chat.ID).Info("Joining chat")
}

// onUserJoined is fired when a user is added (or joins) to a group.
func (bot *telegramBot) onUserJoined(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}

	// If it's me (the bot) joining the chat, don't do anything else
	// Note: this should be replaced with onAddedToGroup
	// TODO: verify if this code is required (might be replaced by onAddedToGroup)
	if m.IsService() && !m.Private() && m.UserJoined.ID == bot.telebot.Me.ID {
		bot.logger.WithField("chatid", m.Chat.ID).Info("Joining chat")
		return
	}

	// Check if the user that's joining is g-lined. If so, ban them and delete
	// the join service message.
	if banned, err := bot.db.IsUserBanned(m.Sender.ID); err == nil && banned {
		bot.banUser(m.Chat, m.Sender, settings, "user g-lined")
		bot.deleteMessage(m, settings, "user g-lined")
		return
	}

	// Check if the user that's joining is CAS banned. If so, do the proper
	// action.
	if bot.cas != nil && bot.cas.IsBanned(m.Sender.ID) {
		bot.casDatabaseMatch.Inc()
		bot.performAction(m, m.Sender, settings, settings.OnBlacklistCAS, "CAS banned")
		return
	}

	// Check for spam items in user names.
	textvalues := []string{
		m.UserJoined.Username,
		m.UserJoined.FirstName,
		m.UserJoined.LastName,
	}
	bot.spamFilter(m, settings, textvalues)

	// If the owner wants to delete all join messages, do so.
	if settings.OnJoinDelete {
		if err := ctx.Delete(); err != nil {
			bot.logger.WithError(err).Error("Failed to delete join message")
		}
	}
}

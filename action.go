package main

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func performAction(m *tb.Message, u *tb.User, action botdatabase.BotAction) {
	switch action.Action {
	case botdatabase.ACTION_MUTE:
		logger.Debugf("Action: mute user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		muteUser(m.Chat, u, m)
		actionDelete(m)
	case botdatabase.ACTION_BAN:
		logger.Debugf("Action: ban user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		banUser(m.Chat, u)
		actionDelete(m)
	case botdatabase.ACTION_DELETE_MSG:
		logger.Debugf("Action: delete message %d of user %d (%s %s %s) in chat %d (%s)", m.ID,
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		actionDelete(m)
	case botdatabase.ACTION_KICK:
		logger.Debugf("Action: kick user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		kickUser(m.Chat, u)
		actionDelete(m)
	case botdatabase.ACTION_NONE:
		logger.Debugf("Action: NONE for user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
	default:
	}
}

func prettyActionName(action botdatabase.BotAction) string {
	switch action.Action {
	case botdatabase.ACTION_MUTE:
		return "🔇 Mute"
	case botdatabase.ACTION_BAN:
		return "🚷 Ban"
	case botdatabase.ACTION_DELETE_MSG:
		return "✂️ Delete"
	case botdatabase.ACTION_KICK:
		return "❗️ Kick"
	case botdatabase.ACTION_NONE:
		return "💤 Do nothing"
	default:
		return "n/a"
	}
}

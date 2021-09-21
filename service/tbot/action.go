package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) performAction(m *tb.Message, u *tb.User, action botdatabase.BotAction) {
	switch action.Action {
	case botdatabase.ACTION_MUTE:
		bot.logger.Debugf("Action: mute user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		bot.muteUser(m.Chat, u, m)
		bot.actionDelete(m)
	case botdatabase.ACTION_BAN:
		bot.logger.Debugf("Action: ban user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		bot.banUser(m.Chat, u)
		bot.actionDelete(m)
	case botdatabase.ACTION_DELETE_MSG:
		bot.logger.Debugf("Action: delete message %d of user %d (%s %s %s) in chat %d (%s)", m.ID,
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		bot.actionDelete(m)
	case botdatabase.ACTION_KICK:
		bot.logger.Debugf("Action: kick user %d (%s %s %s) in chat %d (%s)",
			u.ID, u.Username, u.FirstName, u.LastName, m.Chat.ID, m.Chat.Title)
		bot.kickUser(m.Chat, u)
		bot.actionDelete(m)
	case botdatabase.ACTION_NONE:
		bot.logger.Debugf("Action: NONE for user %d (%s %s %s) in chat %d (%s)",
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

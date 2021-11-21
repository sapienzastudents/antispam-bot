package tbot

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v3"
)

// performAction is a multiplexer function used to do an action (muteUser, banUser, kickUser, deleteMessage) based on
// the chat settings
func (bot *telegramBot) performAction(message *tb.Message, user *tb.User, settings chatSettings, action botdatabase.BotAction, reason string) {
	switch action.Action {
	case botdatabase.ActionMute:
		bot.muteUser(message.Chat, user, settings, reason)
		bot.deleteMessage(message, settings, reason)
	case botdatabase.ActionBan:
		bot.banUser(message.Chat, user, settings, reason)
		bot.deleteMessage(message, settings, reason)
	case botdatabase.ActionKick:
		bot.kickUser(message.Chat, user, settings, reason)
		bot.deleteMessage(message, settings, reason)
	case botdatabase.ActionDeleteMsg:
		bot.deleteMessage(message, settings, reason)
	case botdatabase.ActionNone:
	default:
	}
}

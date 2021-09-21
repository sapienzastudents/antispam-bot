package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// Useful to ban an user
func (bot *telegramBot) banUser(chat *tb.Chat, user *tb.User) bool {
	chatsettings, err := bot.db.GetChatSetting(bot.telebot, chat)
	if err != nil {
		bot.logger.WithError(err).Error("error getting chat settings")
		return false
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(user) {
		return false
	}

	member, err := bot.telebot.ChatMemberOf(chat, user)
	if err != nil {
		bot.logger.WithError(err).Errorf("Cannot get the member object for user %s (%s %s) in chat %s",
			user.Username, user.FirstName, user.LastName, chat.Title)
	} else {
		err = bot.telebot.Ban(chat, member)
		if err != nil {
			bot.logger.WithError(err).Errorf("Cannot get the member object for user %s (%s %s) in chat %s",
				user.Username, user.FirstName, user.LastName, chat.Title)
			return false
		}
		bot.logger.Infof("User %d banned", user.ID)
		return true
	}
	return false
}

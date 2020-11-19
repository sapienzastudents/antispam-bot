package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// Useful to ban an user
func banUser(chat *tb.Chat, user *tb.User) bool {
	chatsettings, err := botdb.GetChatSetting(b, chat)
	if err != nil {
		logger.WithError(err).Error("error getting chat settings")
		return false
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(user) {
		return false
	}

	member, err := b.ChatMemberOf(chat, user)
	if err != nil {
		logger.WithError(err).Errorf("Cannot get the member object for user %s (%s %s) in chat %s",
			user.Username, user.FirstName, user.LastName, chat.Title)
	} else {
		err = b.Ban(chat, member)
		if err != nil {
			logger.WithError(err).Errorf("Cannot get the member object for user %s (%s %s) in chat %s",
				user.Username, user.FirstName, user.LastName, chat.Title)
			return false
		}
		logger.Infof("User %d banned", user.ID)
		return true
	}
	return false
}

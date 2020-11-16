package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// Useful to kick an user
func kickUser(chat *tb.Chat, user *tb.User) bool {
	chatsettings, err := botdb.GetChatSetting(chat)
	if err != nil {
		logger.Critical(err)
		return false
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(user) {
		return false
	}

	member, err := b.ChatMemberOf(chat, user)
	if err != nil {
		logger.Criticalf("Cannot get the member object for user %s (%s %s) in chat %s %s: %s",
			user.Username, user.FirstName, user.LastName, chat.Title, err.Error())
	} else {
		err = b.Ban(chat, member)
		if err != nil {
			logger.Criticalf("Cannot ban member user %s (%s %s) in chat %s %s: %s",
				user.Username, user.FirstName, user.LastName, chat.Title, err.Error())
		} else {
			_ = b.Unban(chat, user)
			logger.Infof("User %d kicked", user.ID)
			return true
		}
	}
	return false
}

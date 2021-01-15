package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func actionDelete(m *tb.Message) bool {
	chatsettings, err := botdb.GetChatSetting(b, m.Chat)
	if err != nil {
		logger.WithError(err).Error("error getting chat settings")
		return false
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(m.Sender) {
		return false
	}

	err = b.Delete(m)
	if err != nil {
		logger.WithError(err).Errorf("Cannot delete message from user %s %s (%s) in chat %s",
			m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Chat.Title)
		return false
	}
	return true
}

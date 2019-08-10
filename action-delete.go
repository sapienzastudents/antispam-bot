package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func actionDelete(m *tb.Message) bool {
	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	isAdmin, err := IsAdminOf(m.Chat, m.Sender)
	if err != nil || isAdmin {
		return false
	}

	err = b.Delete(m)
	if err != nil {
		logger.Criticalf("Cannot delete message from user %s %s (%s) in chat %s (%d): %s",
			m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Chat.Title, m.Chat.ID, err.Error())
		return false
	} else {
		return true
	}
}

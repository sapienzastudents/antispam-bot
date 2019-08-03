package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func actionDelete(m *tb.Message) {
	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	// TODO: cache this list as it might be slow to look up every time
	/*
		admins, err := b.AdminsOf(m.Chat)
		if err != nil {
			logger.Criticalf("Cannot get the admin list for %s (%d): %s", m.Chat.Title, m.Chat.ID, err.Error())
			return
		}
		for _, a := range admins {
			if m.Sender.ID == a.User.ID {
				logger.Infof("Ok we were wrong, %s %s (%s) is an admin. I can't delete a message from an admin!",
					m.Sender.FirstName, m.Sender.LastName, m.Sender.Username)
				return
			}
		}*/

	err := b.Delete(m)
	if err != nil {
		logger.Criticalf("Cannot delete message from user %s %s (%s) in chat %s (%d): %s",
			m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Chat.Title, m.Chat.ID, err.Error())
	}
}

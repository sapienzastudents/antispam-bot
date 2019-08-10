package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func IsAdminOf(chat *tb.Chat, user *tb.User) (bool, error) {
	// TODO: cache this list as it might be slow to look up every time
	admins, err := b.AdminsOf(chat)
	if err != nil {
		logger.Criticalf("Cannot get the admin list for %s (%d): %s", chat.Title, chat.ID, err.Error())
		return false, err
	}
	for _, a := range admins {
		if user.ID == a.User.ID {
			logger.Infof("Ok we were wrong, %s %s (%s) is an admin. I can't delete a message from an admin!",
				a.User.FirstName, a.User.LastName, a.User.Username)
			return false, nil
		}
	}
	return true, nil
}

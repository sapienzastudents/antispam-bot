package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Useful to mute an user
func muteUser(chat *tb.Chat, user *tb.User, message *tb.Message) bool {
	chatsettings, err := botdb.GetChatSetting(b, message.Chat)
	if err != nil {
		logger.WithError(err).Error("error getting chat settings")
		return false
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(message.Sender) {
		return false
	}

	member, err := b.ChatMemberOf(chat, user)
	if err != nil {
		logger.WithError(err).Errorf("Cannot get the member object for user %s (%s %s) in chat %s",
			user.Username, user.FirstName, user.LastName, chat.Title)
	} else {
		member.CanSendMedia = false
		member.CanSendMessages = false
		member.CanSendOther = false
		err = b.Restrict(chat, member)
		if err != nil {
			logger.WithError(err).Errorf("Cannot save member restriction for user %s (%s %s) in chat %s",
				user.Username, user.FirstName, user.LastName, chat.Title)
		} else if message != nil {
			// If last parameter is a system message, reply to it, otherwise don't say anything

			displayName := "@" + user.Username
			if displayName == "@" {
				// No username set
				displayName = user.FirstName + " " + user.LastName
			}
			_, err := b.Send(chat, fmt.Sprintf("Oh no %s! My SPAM algorithm was triggered and I muted you from the chat.\n\n"+
				"Please, send me a private message so I can unblock you", displayName), &tb.SendOptions{
				DisableNotification: true,
				ParseMode:           tb.ModeMarkdown,
				ReplyTo:             message,
			})
			if err != nil {
				logger.WithError(err).Errorf("Cannot send mute warning to %s (%s %s) in chat %s",
					user.Username, user.FirstName, user.LastName, chat.Title)
			}
			return true
		} else {
			return true
		}
	}
	return false
}

package tbot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Useful to mute an user
func (bot *telegramBot) muteUser(chat *tb.Chat, user *tb.User, chatsettings chatSettings, reason string) {
	logfields := logrus.Fields{
		"userid": user.ID,
		"chatid": chat.ID,
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(user) {
		return
	}

	member, err := bot.telebot.ChatMemberOf(chat, user)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("mute action: cannot get member object for user")
		return
	}

	member.CanSendMedia = false
	member.CanSendMessages = false
	member.CanSendOther = false
	err = bot.telebot.Restrict(chat, member)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("mute action: cannot save member restriction")
		return
	}

	bot.logger.WithFields(logfields).WithField("reason", reason).Info("mute user")
	chatsettings.Log("mute", bot.telebot.Me, user, reason)
}

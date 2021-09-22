package tbot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Useful to kick an user
func (bot *telegramBot) kickUser(chat *tb.Chat, user *tb.User, chatsettings chatSettings, reason string) {
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
		bot.logger.WithError(err).WithFields(logfields).Error("kick action: cannot get member object for user")
		return
	}

	err = bot.telebot.Ban(chat, member)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("kick action: cannot ban user")
		return
	}

	err = bot.telebot.Unban(chat, user)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("kick action: cannot unban user")
		return
	}
	bot.logger.WithFields(logfields).WithField("reason", reason).Info("kick user")
	chatsettings.Log("kick", bot.telebot.Me, user, reason)
}

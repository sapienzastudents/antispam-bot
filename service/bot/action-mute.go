package bot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

// muteUser mutes the given user on the given chat.
//
// It has no effect on chat admins. It records the action in the log.
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
		bot.logger.WithError(err).WithFields(logfields).Error("mute action: failed to get member object for user")
		return
	}

	member.CanSendMedia = false
	member.CanSendMessages = false
	member.CanSendOther = false
	err = bot.telebot.Restrict(chat, member)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("mute action: failed to save member restriction")
		return
	}

	bot.logger.WithFields(logfields).WithField("reason", reason).Info("mute user")
	chatsettings.Log("mute", bot.telebot.Me, user, reason)
}

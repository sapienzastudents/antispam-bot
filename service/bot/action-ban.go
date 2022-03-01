package bot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

// banUser will ban a user. It has no effect on chat admins. It records the action in the log
func (bot *telegramBot) banUser(chat *tb.Chat, user *tb.User, chatsettings chatSettings, reason string) {
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
		bot.logger.WithError(err).WithFields(logfields).Error("ban action: cannot get member object for user")
		return
	}

	err = bot.telebot.Ban(chat, member)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("ban action: cannot ban user")
		return
	}

	bot.logger.WithFields(logfields).WithField("reason", reason).Info("ban user")
	chatsettings.Log("ban", bot.telebot.Me, user, reason)
}

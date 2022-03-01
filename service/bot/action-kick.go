package bot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

// kickUser kicks the given user on the given chat.
//
// It has no effect on chat admins. It records the action in the log.
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
		bot.logger.WithError(err).WithFields(logfields).Error("kick action: failed to get member object for user")
		return
	}

	// There is no method for kicking a user in Telegram. Banning and un-banning
	// a user is the way to kick a user from the chat.
	err = bot.telebot.Ban(chat, member)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("kick action: failed to ban user")
		return
	}

	err = bot.telebot.Unban(chat, user)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("kick action: failed to unban user")
		return
	}
	bot.logger.WithFields(logfields).WithField("reason", reason).Info("kick user")
	chatsettings.Log("kick", bot.telebot.Me, user, reason)
}

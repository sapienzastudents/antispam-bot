package tbot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

// deleteMessage is useful when deleting a message that needs to be recorded in the log (e.g. a non-system message). It
// has no effect on admins
func (bot *telegramBot) deleteMessage(m *tb.Message, chatsettings chatSettings, reason string) {
	logfields := logrus.Fields{
		"userid":    m.Sender.ID,
		"chatid":    m.Chat.ID,
		"messageid": m.ID,
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if chatsettings.ChatAdmins.IsAdmin(m.Sender) {
		return
	}

	chatsettings.Log("delete message", bot.telebot.Me, m.Sender, reason)
	chatsettings.LogForward(m)

	err := bot.telebot.Delete(m)
	if err != nil {
		bot.logger.WithError(err).WithFields(logfields).Error("delete msg action: can't delete message")
	}
	bot.logger.WithFields(logfields).WithField("reason", reason).Info("delete message")
}

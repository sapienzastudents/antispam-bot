package bot

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v3"
)

// onTerminate terminates the user that the reply /terminate command refers to.
//
// It first warn the user, then starts a contdown of 60 seconds and there is no
// way to stop the timer.
func (bot *telegramBot) onTerminate(ctx tb.Context, settings chatSettings) {
	bot.botCommandsRequestsTotal.WithLabelValues("terminate").Inc()

	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}
	_ = ctx.Delete()

	if !m.IsReply() || m.Private() {
		// No sense when no user is quoted (in public groups) or if the message
		// is sent in private...
		return
	}

	// Do not terminate an admin (remember: The Admin Is Always Right®).
	isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.ReplyTo.Sender.ID)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
		return
	}
	if settings.ChatAdmins.IsAdmin(m.ReplyTo.Sender) || isGlobalAdmin {
		return
	}

	bot.logger.WithFields(logrus.Fields{
		"adminid":        m.Sender.ID,
		"adminusername":  m.Sender.Username,
		"adminfirstname": m.Sender.FirstName,
		"adminlastname":  m.Sender.LastName,
		"user":           m.ReplyTo.Sender.ID,
		"userusername":   m.ReplyTo.Sender.Username,
		"userfirstname":  m.ReplyTo.Sender.FirstName,
		"userlastname":   m.ReplyTo.Sender.LastName,
	}).Debug("User going to be terminated")

	lang := m.Sender.LanguageCode
	if m.Sender.Username != "" {
		_, _ = bot.telebot.Reply(m.ReplyTo, fmt.Sprintf("🚨 @%s "+bot.bundle.T(lang, "You will be terminated in 60 seconds, there will be no further warnings"), m.ReplyTo.Sender.Username))
	} else {
		_, _ = bot.telebot.Reply(m.ReplyTo, fmt.Sprintf("🚨 %s %s "+bot.bundle.T(lang, "You will be terminated in 60 seconds, there will be no further warnings"), m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName))
	}

	go func() {
		time.Sleep(60 * time.Second)

		member, err := bot.telebot.ChatMemberOf(m.Chat, m.ReplyTo.Sender)
		if err != nil {
			bot.logger.WithError(err).WithField("userid", m.ReplyTo.ID).Error("Failed to ban user")
			return
		}

		_ = bot.telebot.Ban(m.Chat, member)
	}()
}

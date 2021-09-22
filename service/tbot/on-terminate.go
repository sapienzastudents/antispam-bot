package tbot

import (
	"fmt"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onTerminate(m *tb.Message, settings chatSettings) {
	bot.botCommandsRequestsTotal.WithLabelValues("terminate").Inc()

	_ = bot.telebot.Delete(m)
	if m.ReplyTo == nil || m.Private() {
		// We need an handle
		return
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if settings.ChatAdmins.IsAdmin(m.ReplyTo.Sender) || bot.db.IsGlobalAdmin(m.ReplyTo.Sender.ID) {
		return
	}

	bot.logger.Debugf("Terminate by %d (%s %s %s) for %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName,
		m.ReplyTo.Sender.ID, m.ReplyTo.Sender.Username, m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName)

	if m.Sender.Username != "" {
		_, _ = bot.telebot.Reply(m.ReplyTo, fmt.Sprintf("🚨 @%s You will be terminated in 60 seconds, there will be no further warnings", m.ReplyTo.Sender.Username))
	} else {
		_, _ = bot.telebot.Reply(m.ReplyTo, fmt.Sprintf("🚨 %s %s You will be terminated in 60 seconds, there will be no further warnings", m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName))
	}

	go func() {
		time.Sleep(60 * time.Second)

		member, err := bot.telebot.ChatMemberOf(m.Chat, m.ReplyTo.Sender)
		if err != nil {
			bot.logger.WithError(err).Error("Can't ban user ", m.ReplyTo.Sender)
			return
		}

		_ = bot.telebot.Ban(m.Chat, member)
	}()
}

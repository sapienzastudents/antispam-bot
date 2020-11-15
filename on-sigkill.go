package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

func onTerminate(m *tb.Message, settings ChatSettings) {
	b.Delete(m)
	if m.ReplyTo == nil || m.Private() {
		// We need an handle
		return
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	isAdmin, err := IsAdminOf(m.Chat, m.ReplyTo.Sender)
	if err != nil || isAdmin {
		if err != nil {
			logger.Critical(err)
		}
		return
	}

	logger.Debugf("Sigkill %d (%s %s %s) for %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName,
		m.ReplyTo.Sender.ID, m.ReplyTo.Sender.Username, m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName)

	_, _ = b.Reply(m.ReplyTo, "🚨 You will be terminated in 60 seconds, there will be no further warnings")

	go func() {
		time.Sleep(60 * time.Second)

		member, err := b.ChatMemberOf(m.Chat, m.ReplyTo.Sender)
		if err != nil {
			logger.Errorf("Can't ban user %d: %s", m.ReplyTo.Sender, err)
			return
		}

		_ = b.Ban(m.Chat, member)
	}()
}

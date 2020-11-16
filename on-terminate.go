package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

func onTerminate(m *tb.Message, settings ChatSettings) {
	_ = b.Delete(m)
	if m.ReplyTo == nil || m.Private() {
		// We need an handle
		return
	}

	// If the user is an admin, be polite (remember: The Admin Is Always Right®)
	if settings.ChatAdmins.IsAdmin(m.ReplyTo.Sender) {
		return
	}

	logger.Debugf("Terminate by %d (%s %s %s) for %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName,
		m.ReplyTo.Sender.ID, m.ReplyTo.Sender.Username, m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName)

	if m.Sender.Username != "" {
		_, _ = b.Reply(m.ReplyTo, fmt.Sprintf("🚨 @%s You will be terminated in 60 seconds, there will be no further warnings", m.ReplyTo.Sender.Username))
	} else {
		_, _ = b.Reply(m.ReplyTo, fmt.Sprintf("🚨 %s %s You will be terminated in 60 seconds, there will be no further warnings", m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName))
	}

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

package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func onMyChatroomRequest(m *tb.Message, _ ChatSettings) {
	logger.Debugf("My chat room requested by %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName)
	if m.Private() && botdb.IsGlobalAdmin(m.Sender) {
		chatrooms, err := botdb.ListMyChatrooms()
		if err != nil {
			logger.Criticalf("Error getting chatroom list: %s", err.Error())
		} else {
			msg := strings.Builder{}
			msg.WriteString("Chatrooms:\n\n")
			for _, v := range chatrooms {
				msg.WriteString("- ")
				msg.WriteString(v.Title)
				msg.WriteString(" (")
				msg.WriteString(string(v.Type))
				msg.WriteString(")\n")
			}
			_, _ = b.Send(m.Chat, msg.String())
		}
	} else {
		logger.Warningf("User %d is not a global admin, chatroom request denied", m.Sender.ID)
	}
}

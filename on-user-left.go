package main

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserLeft(m *tb.Message, settings botdatabase.ChatSettings) {
	logger.Infof("User %d left chat %s (%d)", m.UserLeft.ID, m.Chat.Title, m.Chat.ID)
	if !m.Private() && m.UserLeft.ID == b.Me.ID {
		_ = botdb.LeftChatroom(m.Chat)
	}
	if settings.OnLeaveDelete {
		err := b.Delete(m)
		if err != nil {
			logger.WithError(err).Error("Cannot delete leave message")
		}
	}
}

package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserLeft(m *tb.Message, settings ChatSettings) {
	logger.Infof("User %d left chat %s (%d)", m.UserLeft.ID, m.Chat.Title, m.Chat.ID)
	if !m.Private() && m.UserLeft.ID == b.Me.ID {
		_ = botdb.LeftChatroom(m.Chat)
	}
	if settings.OnLeaveDelete {
		err := b.Delete(m)
		if err != nil {
			logger.Critical("Cannot delete leave message: ", err)
		}
	}
}

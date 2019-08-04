package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserJoined(m *tb.Message) {
	if b, err := botdb.IsBotEnabled(m.Chat); b || err != nil {
		return
	}
	if m.IsService() && !m.Private() && m.UserJoined.ID == b.Me.ID {
		logger.Infof("Joining chat %s", m.Chat.Title)
	}
	if m.IsService() && !m.Private() {
		botdb.UpdateMyChatroomList(m.Chat)
	}
	if m.IsService() && !m.Private() && m.UserJoined.ID != b.Me.ID {
		// We can mute users from the beginning
		// TODO: leave as an option for admins
		/*if m.UserJoined.Username == "" && m.UserJoined.FirstName == "" && m.UserJoined.LastName == "" {
			muteUser(m.Chat, m.UserJoined, m)
		} else if chineseChars(m.UserJoined.FirstName) > 0.5 || chineseChars(m.UserJoined.LastName) > 0.5 {
			muteUser(m.Chat, m.UserJoined, m)
		}*/
	}
}

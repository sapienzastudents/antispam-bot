package main

import (
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onHelp(m *tb.Message, _ botdatabase.ChatSettings) {
	if m.Private() {
		// === GROUPS button
		var buttons [][]tb.InlineButton
		var groupsBt = tb.InlineButton{
			Unique: "bt_action_groups",
			Text:   "🇬🇧 Groups / 🇮🇹 Gruppi",
		}
		b.Handle(&groupsBt, func(callback *tb.Callback) {
			_ = b.Respond(callback)
			// Note that the second parameter is always ignored in onGroups, so we can avoid a db lookup
			onGroups(m, botdatabase.ChatSettings{})
		})
		buttons = append(buttons, []tb.InlineButton{groupsBt})

		// === SETTINGS button
		// Check if the user is an admin in at least one chat
		var settingsVisible = false
		if !botdb.IsGlobalAdmin(m.Sender) {
			chatrooms, err := botdb.ListMyChatrooms()
			if err != nil {
				logger.WithError(err).Error("cant get chatroom list")
			} else {
				for _, x := range chatrooms {
					chatsettings, err := botdb.GetChatSetting(b, x)
					if err != nil {
						logger.WithError(err).WithField("chat", x.ID).Warn("can't get chatroom settings")
						continue
					}
					if chatsettings.ChatAdmins.IsAdmin(m.Sender) {
						settingsVisible = true
						break
					}
				}
			}
		} else {
			settingsVisible = true
		}

		if settingsVisible {
			var settingsBt = tb.InlineButton{
				Unique: "bt_action_settings",
				Text:   "🇬🇧 Settings / 🇮🇹 Impostazioni",
			}
			b.Handle(&settingsBt, func(callback *tb.Callback) {
				_ = b.Respond(callback)
				// Note that the second parameter is always ignored in onSettings when asking from a direct message, so we
				// can avoid a db lookup
				onSettings(m, botdatabase.ChatSettings{})
			})
			buttons = append(buttons, []tb.InlineButton{settingsBt})
		}

		_, err := b.Send(m.Sender, "🇮🇹 Ciao! Cosa cerchi?\n\n🇬🇧 Hi! What are you looking for?", &tb.SendOptions{
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: buttons,
			},
		})
		if err != nil {
			logger.WithError(err).Warn("can't send message on help")
		}
	}
}

package main

import (
	"strings"

	"github.com/google/uuid"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func startFromUUID(payload string, sender *tb.User) {
	chatUUID, err := uuid.Parse(payload)
	if err != nil {
		logger.WithError(err).Error("error parsing chat UUID")
		return
	}

	chatID, err := botdb.GetChatIDFromUUID(chatUUID)
	if err != nil {
		logger.WithError(err).Error("chat not found for UUID " + chatUUID.String())
		return
	}

	var msg string

	inviteLink, err := getInviteLink(&tb.Chat{ID: chatID})
	if err != nil {
		logger.WithError(err).WithField("chat", chatID).Warning("can't generate invite link")
		msg = "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-("
	} else {
		msg = "🇮🇹 Ciao! Il link di invito è questo qui sotto (se dice che non è funzionante, riprova ad usarlo tra 1-2 minuti):\n\n🇬🇧 Hi! The invite link is the following (if Telegram says that it's invalid, wait 1-2 minutes before using it):\n\n" + inviteLink
	}

	_, err = b.Send(sender, msg)
	if err != nil {
		logger.WithError(err).Warn("can't send message on help from web")
	}
}

func onHelp(m *tb.Message, _ botdatabase.ChatSettings) {
	botCommandsRequestsTotal.WithLabelValues("start").Inc()
	_ = b.Delete(m)

	if m.Private() {
		payload := strings.TrimSpace(m.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(m.Text, " ")
			startFromUUID(parts[1], m.Sender)
			return
		}

		// === GROUPS button
		var buttons [][]tb.InlineButton
		var groupsBt = tb.InlineButton{
			Unique: "bt_action_groups",
			Text:   "🇬🇧 Groups / 🇮🇹 Gruppi",
		}
		b.Handle(&groupsBt, func(callback *tb.Callback) {
			_ = b.Respond(callback)
			// Note that the second parameter is always ignored in onGroups, so we can avoid a db lookup
			sendGroupListForLinks(callback.Sender, callback.Message, callback.Message.Chat, nil)
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
				sendGroupListForSettings(callback.Sender, callback.Message, callback.Message.Chat, 0)
			})
			buttons = append(buttons, []tb.InlineButton{settingsBt})
		}

		var bt = tb.InlineButton{
			Unique: "help_close",
			Text:   "Close / Chiudi",
		}
		buttons = append(buttons, []tb.InlineButton{bt})
		b.Handle(&bt, func(callback *tb.Callback) {
			_ = b.Respond(callback)
			_ = b.Delete(callback.Message)
		})

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

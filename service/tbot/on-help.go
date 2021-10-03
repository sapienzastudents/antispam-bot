package tbot

import (
	"strings"

	"github.com/google/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

// startFromUUID replies to the user with the invite link of a chat from a UUID. The UUID is used in the website to
// avoid SPAM bots. All links in the website will cause the user to send a "/start UUID" message
func (bot *telegramBot) startFromUUID(payload string, sender *tb.User) {
	chatUUID, err := uuid.Parse(payload)
	if err != nil {
		bot.logger.WithError(err).Error("error parsing chat UUID")
		return
	}

	chatID, err := bot.db.GetChatIDFromUUID(chatUUID)
	if err != nil {
		bot.logger.WithError(err).Error("chat not found for UUID " + chatUUID.String())
		return
	}

	var msg string

	inviteLink, err := bot.getInviteLink(&tb.Chat{ID: chatID})
	if err != nil {
		bot.logger.WithError(err).WithField("chat", chatID).Warning("can't generate invite link")
		msg = "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-("
	} else {
		msg = "🇮🇹 Ciao! Il link di invito è questo qui sotto (se dice che non è funzionante, riprova ad usarlo tra 1-2 minuti):\n\n🇬🇧 Hi! The invite link is the following (if Telegram says that it's invalid, wait 1-2 minutes before using it):\n\n" + inviteLink
	}

	_, err = bot.telebot.Send(sender, msg)
	if err != nil {
		bot.logger.WithError(err).Warn("can't send message on help from web")
	}
}

// onHelp replies with a tiny help text, and [Groups] and [Settings] buttons. It replies also for /start
// When /start is followed by a UUID, then calls startFromUUID to handle the UUID<->chat mapping
func (bot *telegramBot) onHelp(m *tb.Message, _ chatSettings) {
	bot.botCommandsRequestsTotal.WithLabelValues("start").Inc()
	_ = bot.telebot.Delete(m)

	if m.Private() {
		// If a UUID is specified, then we need to lookup for the invite link and send it to the user
		payload := strings.TrimSpace(m.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(m.Text, " ")
			bot.startFromUUID(parts[1], m.Sender)
			return
		}

		// === GROUPS button
		var buttons [][]tb.InlineButton
		var groupsBt = tb.InlineButton{
			Unique: "bt_action_groups",
			Text:   "🇬🇧 Groups / 🇮🇹 Gruppi",
		}
		bot.telebot.Handle(&groupsBt, func(callback *tb.Callback) {
			_ = bot.telebot.Respond(callback)
			// Note that the second parameter is always ignored in onGroups, so we can avoid a db lookup
			bot.sendGroupListForLinks(callback.Sender, callback.Message, callback.Message.Chat, nil)
		})
		buttons = append(buttons, []tb.InlineButton{groupsBt})

		isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("can't check if the user is a global admin")
			return
		}

		// === SETTINGS button
		// Check if the user is an admin in at least one chat
		var settingsVisible = false
		if !isGlobalAdmin {
			chatrooms, err := bot.db.ListMyChatrooms()
			if err != nil {
				bot.logger.WithError(err).Error("cant get chatroom list")
			} else {
				for _, x := range chatrooms {
					chatsettings, err := bot.getChatSettings(x)
					if err != nil {
						bot.logger.WithError(err).WithField("chat", x.ID).Warn("can't get chatroom settings")
						continue
					}
					if chatsettings.ChatAdmins.IsAdmin(m.Sender) {
						settingsVisible = true
						break
					}
				}
			}
		} else {
			// The global bot admin is always able to see group settings button
			settingsVisible = true
		}

		if settingsVisible {
			var settingsBt = tb.InlineButton{
				Unique: "bt_action_settings",
				Text:   "🇬🇧 Settings / 🇮🇹 Impostazioni",
			}
			bot.telebot.Handle(&settingsBt, func(callback *tb.Callback) {
				_ = bot.telebot.Respond(callback)
				bot.sendGroupListForSettings(callback.Sender, callback.Message, callback.Message.Chat, 0)
			})
			buttons = append(buttons, []tb.InlineButton{settingsBt})
		}

		/*
			guida per aggiungere il bot
		*/
		var guidebt = tb.InlineButton{
			Unique: "guide",
			Text:   "⚙️ Come aggiungere un gruppo",
		}

		bot.telebot.Handle(&guidebt, func(callback *tb.Callback) {
			_ = bot.telebot.Respond(callback)
			// Note that the second parameter is always ignored in onGroups, so we can avoid a db lookup
			bot.onGuide(callback.Message)
		})

		buttons = append(buttons, []tb.InlineButton{guidebt})

		// === CLOSE Button
		var bt = tb.InlineButton{
			Unique: "help_close",
			Text:   "Close / Chiudi",
		}
		buttons = append(buttons, []tb.InlineButton{bt})
		bot.telebot.Handle(&bt, func(callback *tb.Callback) {
			_ = bot.telebot.Respond(callback)
			_ = bot.telebot.Delete(callback.Message)
		})

		_, err = bot.telebot.Send(m.Sender, "🇮🇹 Ciao! Cosa cerchi?\n\n🇬🇧 Hi! What are you looking for?", &tb.SendOptions{
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: buttons,
			},
		})
		if err != nil {
			bot.logger.WithError(err).Warn("can't send message on help")
		}
	}
}

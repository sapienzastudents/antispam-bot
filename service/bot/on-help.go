package bot

import (
	"strings"

	"github.com/google/uuid"
	tb "gopkg.in/telebot.v3"
)

// startFromUUID replies to the user with the invite link of a chat from a UUID.
// The UUID is used in the website to avoid SPAM bots. All links in the website
// will cause the user to send a "/start UUID" message
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
	lang := sender.LanguageCode

	inviteLink, err := bot.getInviteLink(&tb.Chat{ID: chatID})
	if err != nil {
		bot.logger.WithError(err).WithField("chat", chatID).Warning("Failed to generate invite link")
		msg = bot.bundle.T(lang, "Oops, I'm broken, please get in touch with my admin!")
	} else {
		msg = bot.bundle.T(lang, "Hi! The invite link is the following (if Telegram says that it's invalid, wait 1-2 minutes before using it):") + "\n\n" + inviteLink
	}

	_, err = bot.telebot.Send(sender, msg)
	if err != nil {
		bot.logger.WithError(err).Warn("Failed to send message on help from web")
	}
}

// on Help command replies on /help with a small help message following two
// buttons, "Groups" and "Settings".
//
// It replies also for /start command. When that command is followed by an UUID,
// then calls startFromUUID to handle the UUID and chat mapping.
func (bot *telegramBot) onHelp(ctx tb.Context, settings chatSettings) {
	msg := ctx.Message()

	if msg.Private() {
		// If a UUID is specified, then we need to lookup for the invite link
		// and send it to the user.
		payload := strings.TrimSpace(msg.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(msg.Text, " ")
			bot.startFromUUID(parts[1], msg.Sender)
			return
		}

		bot.sendHelpMessage(ctx.Sender(), nil)
	}
}

// sendHelpMessage sends the help message to the given user.
//
// If message is not nil, it edits this message instead of sending a new one.
func (bot *telegramBot) sendHelpMessage(user *tb.User, message *tb.Message) {
	// IETF language tag used to localize messages.
	lang := user.LanguageCode

	var buttons [][]tb.InlineButton

	// "Groups" button.
	groupsBt := tb.InlineButton{
		Unique: "bt_action_groups",
		Text:   "🏘 " + bot.bundle.T(lang, "Groups"),
	}
	bot.telebot.Handle(&groupsBt, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}

		// Note that the second parameter is always ignored in onGroups, so
		// we can avoid a DB lookup.
		bot.sendGroupListForLinks(ctx.Sender(), ctx.Message(), ctx.Message().Chat, nil)
		return nil
	})
	buttons = append(buttons, []tb.InlineButton{groupsBt})

	isGlobalAdmin, err := bot.db.IsBotAdmin(user.ID)
	if err != nil {
		bot.logger.WithField("user_id", user.ID).WithError(err).Error("Failed to check if the user is a global admin")
		return
	}

	// "Settings" button.
	// Check if the user is an admin in at least one chat.
	settingsVisible := false
	if !isGlobalAdmin {
		chatrooms, err := bot.db.ListMyChats()
		if err != nil {
			bot.logger.WithError(err).Error("Failed to get chatroom list")
			return
		}
		for _, x := range chatrooms {
			chatsettings, err := bot.getChatSettings(x)
			if err != nil {
				bot.logger.WithError(err).WithField("chatid", x.ID).Warn("Failed to get chatroom settings")
				continue
			}
			if chatsettings.ChatAdmins.IsAdmin(user) {
				settingsVisible = true
				break
			}
		}
	} else {
		// The global bot admin is always able to see group settings button.
		settingsVisible = true
	}

	if settingsVisible {
		settingsBt := tb.InlineButton{
			Unique: "bt_action_settings",
			Text:   "⚙️ " + bot.bundle.T(lang, "Settings"),
		}
		bot.telebot.Handle(&settingsBt, func(ctx tb.Context) error {
			if err := ctx.Respond(); err != nil {
				bot.logger.WithError(err).Error("Failed to respond to callback query")
				return err
			}
			bot.sendGroupListForSettings(ctx.Sender(), ctx.Message(), ctx.Message().Chat, 0)
			return nil
		})
		buttons = append(buttons, []tb.InlineButton{settingsBt})
	}

	// "Blacklist" button.
	blacklistBt := tb.InlineButton{
		Unique: "bt_action_blacklist",
		Text:   "⚫️ " + bot.bundle.T(lang, "Blacklist"),
	}
	bot.telebot.Handle(&blacklistBt, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}

		bot.sendBlacklist(ctx.Sender(), ctx.Message(), 0)
		return nil
	})
	buttons = append(buttons, []tb.InlineButton{blacklistBt})

	if isGlobalAdmin {
		adminSettingsBt := tb.InlineButton{
			Unique: "bt_action_admin_settings",
			Text:   "👮" + bot.bundle.T(lang, "Admins settings"),
		}
		bot.telebot.Handle(&adminSettingsBt, func(ctx tb.Context) error {
			if err := ctx.Respond(); err != nil {
				bot.logger.WithError(err).Error("Failed to respond to callback query")
				return err
			}
			bot.sendAdminsForSettings(ctx.Sender(), ctx.Message())
			return nil
		})
		buttons = append(buttons, []tb.InlineButton{adminSettingsBt})
	}

	contactsbt := tb.InlineButton{
		Unique: "contacts",
		Text:   "❓ " + bot.bundle.T(lang, "Contacts"),
	}
	bot.telebot.Handle(&contactsbt, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}
		return bot.onContacts(ctx)
	})
	buttons = append(buttons, []tb.InlineButton{contactsbt})

	// Help button, used to show a small help message on how to add the bot
	// on a group.
	guidebt := tb.InlineButton{
		Unique: "guide",
		Text:   "ℹ️  " + bot.bundle.T(lang, "How to add a group"),
	}
	bot.telebot.Handle(&guidebt, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}
		bot.onGuide(ctx)
		return nil
	})
	buttons = append(buttons, []tb.InlineButton{guidebt})

	// Close button.
	bt := tb.InlineButton{
		Unique: "help_close",
		Text:   "🚪 " + bot.bundle.T(lang, "Close"),
	}
	buttons = append(buttons, []tb.InlineButton{bt})
	bot.telebot.Handle(&bt, func(ctx tb.Context) error {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return err
		}
		return bot.telebot.Delete(ctx.Callback().Message)
	})

	msg := "👋 " + bot.bundle.T(lang, "Hi! What are you looking for?")
	sendOptions := &tb.ReplyMarkup{InlineKeyboard: buttons}
	if message == nil {
		_, err = bot.telebot.Send(user, msg, sendOptions)
	} else {
		_, err = bot.telebot.Edit(message, msg, sendOptions)
	}
	if err != nil {
		bot.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to send/edit message for chat")
	}
}

package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
	"time"
)

func generateSettingsMessageText(chat *tb.Chat, settings ChatSettings) string {
	reply := strings.Builder{}

	reply.WriteString(fmt.Sprintf("Bot settings for chat %s (%d)\n\n", chat.Title, chat.ID))
	if settings.BotEnabled {
		reply.WriteString("✅ Bot enabled\n")
	} else {
		reply.WriteString("❌ Bot disabled\n")
	}

	if !settings.Hidden {
		reply.WriteString("👀 Group visible\n")
	} else {
		reply.WriteString("⛔️ Group hidden\n")
	}

	if settings.OnJoinDelete {
		reply.WriteString("✅ Delete join message (after spam detection)\n")
	} else {
		reply.WriteString("❌ Do not delete join messages (after spam detection)\n")
	}

	if settings.OnLeaveDelete {
		reply.WriteString("✅ Delete leave message\n")
	} else {
		reply.WriteString("❌ Do not delete leave messages\n")
	}

	reply.WriteString("\n🇨🇳 *Chinese* blocker:\n")
	reply.WriteString("On join: *")
	reply.WriteString(prettyActionName(settings.OnJoinChinese))
	reply.WriteString("*\n")
	reply.WriteString("On message: *")
	reply.WriteString(prettyActionName(settings.OnMessageChinese))
	reply.WriteString("*\n")

	reply.WriteString("\n☪️ *Arabic* blocker:\n")
	reply.WriteString("On join: *")
	reply.WriteString(prettyActionName(settings.OnJoinArabic))
	reply.WriteString("*\n")
	reply.WriteString("On message: *")
	reply.WriteString(prettyActionName(settings.OnMessageArabic))
	reply.WriteString("*\n")

	reply.WriteString("\nCAS-ban (see https://combot.org/cas/ ):\n")
	reply.WriteString("On any action: *")
	reply.WriteString(prettyActionName(settings.OnBlacklistCAS))
	reply.WriteString("*\n")

	reply.WriteString("\nGenerated at: ")
	reply.WriteString(time.Now().String())
	return reply.String()
}

func generateSettingsReplyMarkup(chat *tb.Chat, settings ChatSettings) *tb.ReplyMarkup {
	// TODO: move button creations in init function (eg. global buttons objects and handler)
	settingsRefreshButton := tb.InlineButton{
		Unique: "settings_message_refresh",
		Text:   "🔄 Refresh",
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&settingsRefreshButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		return settings
	}))

	// Enable / Disable bot button
	enableDisableButtonText := "✅ Enable bot"
	if settings.BotEnabled {
		enableDisableButtonText = "❌ Disable bot"
	}
	enableDisableBotButton := tb.InlineButton{
		Unique: "settings_enable_disable_bot",
		Text:   enableDisableButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&enableDisableBotButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		settings.BotEnabled = !settings.BotEnabled
		return settings
	}))

	// Hide / show group in group lista
	hideShowButtonText := "👀 Show group"
	if !settings.Hidden {
		hideShowButtonText = "⛔️ Hide group"
	}
	hideShowBotButton := tb.InlineButton{
		Unique: "settings_show_hide_group",
		Text:   hideShowButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&hideShowBotButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		settings.Hidden = !settings.Hidden
		return settings
	}))

	// Delete join and part messages
	deleteJoinMessagesText := "✅ Del join msgs"
	if settings.OnJoinDelete {
		deleteJoinMessagesText = "❌ Don't del join msgs"
	}
	deleteJoinMessages := tb.InlineButton{
		Unique: "settings_enable_disable_delete_on_join",
		Text:   deleteJoinMessagesText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&deleteJoinMessages, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		settings.OnJoinDelete = !settings.OnJoinDelete
		return settings
	}))

	deleteLeaveMessagesText := "✅ Del leave msgs"
	if settings.OnLeaveDelete {
		deleteLeaveMessagesText = "❌ Don't del leave msgs"
	}
	deleteLeaveMessages := tb.InlineButton{
		Unique: "settings_enable_disable_delete_on_leave",
		Text:   deleteLeaveMessagesText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&deleteLeaveMessages, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		settings.OnLeaveDelete = !settings.OnLeaveDelete
		return settings
	}))

	// On Join Chinese (TODO: add kick action)
	onJoinChineseKickButtonText := "✅ Ban Chinese on join"
	if settings.OnJoinChinese.Action != ACTION_NONE {
		onJoinChineseKickButtonText = "❌ Don't ban chinese joins"
	}
	onJoinChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_join",
		Text:   onJoinChineseKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onJoinChineseKickButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		if settings.OnJoinChinese.Action == ACTION_NONE {
			settings.OnJoinChinese = BotAction{
				Action: ACTION_BAN,
			}
		} else {
			settings.OnJoinChinese = BotAction{
				Action: ACTION_NONE,
			}
		}
		return settings
	}))

	// On Join Arabic (TODO: add kick action)
	onJoinArabicKickButtonText := "✅ Ban Arabic on join"
	if settings.OnJoinArabic.Action != ACTION_NONE {
		onJoinArabicKickButtonText = "❌ Don't ban arabs joins"
	}
	onJoinArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_join",
		Text:   onJoinArabicKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onJoinArabicKickButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		if settings.OnJoinArabic.Action == ACTION_NONE {
			settings.OnJoinArabic = BotAction{
				Action: ACTION_BAN,
			}
		} else {
			settings.OnJoinArabic = BotAction{
				Action: ACTION_NONE,
			}
		}
		return settings
	}))

	// On Message Chinese (TODO: add ban action)
	onMessageChineseKickButtonText := "✅ Kick Chinese msgs"
	if settings.OnMessageChinese.Action != ACTION_NONE {
		onMessageChineseKickButtonText = "❌ Don't kick chinese msgs"
	}
	onMessageChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_msgs",
		Text:   onMessageChineseKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onMessageChineseKickButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		if settings.OnMessageChinese.Action == ACTION_NONE {
			settings.OnMessageChinese = BotAction{
				Action: ACTION_KICK,
			}
		} else {
			settings.OnMessageChinese = BotAction{
				Action: ACTION_NONE,
			}
		}
		return settings
	}))

	// On Message Arabic (TODO: add ban action)
	onMessageArabicKickButtonText := "✅ Kick Arabic msgs"
	if settings.OnMessageArabic.Action != ACTION_NONE {
		onMessageArabicKickButtonText = "❌ Don't kick arabs msgs"
	}
	onMessageArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_msgs",
		Text:   onMessageArabicKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onMessageArabicKickButton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		if settings.OnMessageArabic.Action == ACTION_NONE {
			settings.OnMessageArabic = BotAction{
				Action: ACTION_KICK,
			}
		} else {
			settings.OnMessageArabic = BotAction{
				Action: ACTION_NONE,
			}
		}
		return settings
	}))

	// Enable CAS
	enableCASbuttonText := "❌ CAS disabled"
	if settings.OnBlacklistCAS.Action != ACTION_NONE {
		enableCASbuttonText = "✅ CAS enabled"
	}
	enableCASbutton := tb.InlineButton{
		Unique: "settings_enable_disable_cas",
		Text:   enableCASbuttonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&enableCASbutton, CallbackSettings(func(callback *tb.Callback, settings ChatSettings) ChatSettings {
		if settings.OnBlacklistCAS.Action == ACTION_NONE {
			settings.OnBlacklistCAS = BotAction{
				Action: ACTION_KICK,
			}
		} else {
			settings.OnBlacklistCAS = BotAction{
				Action: ACTION_NONE,
			}
		}
		return settings
	}))

	closeBtn := tb.InlineButton{
		Unique: "settings_close",
		Text:   "Close",
	}
	b.Handle(&closeBtn, func(callback *tb.Callback) {
		b.Delete(callback.Message)
	})

	return &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{settingsRefreshButton, enableDisableBotButton},
			{hideShowBotButton},
			{deleteJoinMessages, deleteLeaveMessages},
			{onJoinChineseKickButton, onJoinArabicKickButton},
			{onMessageChineseKickButton, onMessageArabicKickButton},
			{enableCASbutton},
			{closeBtn},
		},
	}
}

func onSettings(m *tb.Message, settings ChatSettings) {
	// Messages in a chat
	if !m.Private() && botdb.IsGlobalAdmin(m.Sender) || settings.ChatAdmins.IsAdmin(m.Sender) {
		_, _ = b.Send(m.Chat, generateSettingsMessageText(m.Chat, settings), &tb.SendOptions{
			ParseMode:             tb.ModeMarkdown,
			ReplyMarkup:           generateSettingsReplyMarkup(m.Chat, settings),
			DisableWebPagePreview: true,
		})
	} else if m.Private() {
		chatButtons := [][]tb.InlineButton{}
		chatrooms, err := botdb.ListMyChatrooms()
		if err != nil {
			logger.Critical(err)
			return
		}

		isGlobalAdmin := botdb.IsGlobalAdmin(m.Sender)

		for _, x := range chatrooms {
			if chatsettings, err := botdb.GetChatSetting(x); !isGlobalAdmin || err != nil || !chatsettings.ChatAdmins.IsAdmin(m.Sender) {
				continue
			}

			btn := tb.InlineButton{
				Unique: fmt.Sprintf("select_chatid_%d", x.ID*-1),
				Text:   x.Title,
				Data:   fmt.Sprintf("%d", x.ID),
			}
			b.Handle(&btn, func(callback *tb.Callback) {
				newchat, _ := b.ChatByID(callback.Data)
				b.Delete(callback.Message)

				settings, _ := botdb.GetChatSetting(newchat)
				_, _ = b.Send(callback.Message.Chat, generateSettingsMessageText(newchat, settings), &tb.SendOptions{
					ParseMode:             tb.ModeMarkdown,
					ReplyMarkup:           generateSettingsReplyMarkup(newchat, settings),
					DisableWebPagePreview: true,
				})
			})
			chatButtons = append(chatButtons, []tb.InlineButton{btn})
		}

		if len(chatButtons) == 0 {
			_, _ = b.Send(m.Chat, "You are not an admin in a chat where the bot is.")
		} else {
			_, _ = b.Send(m.Chat, "Please select the chatroom:", &tb.SendOptions{
				ParseMode: tb.ModeMarkdown,
				ReplyMarkup: &tb.ReplyMarkup{
					InlineKeyboard: chatButtons,
				},
			})
		}
	}
}

func CallbackSettings(fn func(*tb.Callback, ChatSettings) ChatSettings) func(callback *tb.Callback) {
	return func(callback *tb.Callback) {
		var err error
		chat := callback.Message.Chat
		if callback.Data != "" {
			chat, err = b.ChatByID(callback.Data)
			if err != nil {
				logger.Critical(err)
				b.Respond(callback, &tb.CallbackResponse{
					Text:      "Internal error",
					ShowAlert: true,
				})
				return
			}
		}

		settings, err := botdb.GetChatSetting(chat)
		if err != nil {
			logger.Critical("Cannot get chat settings:", err)
			b.Respond(callback, &tb.CallbackResponse{
				Text:      "Internal error",
				ShowAlert: true,
			})
		} else if !settings.ChatAdmins.IsAdmin(callback.Sender) {
			logger.Critical("Non-admin is using a callback from the admin:", callback.Sender)
			b.Respond(callback, &tb.CallbackResponse{
				Text:      "Not authorized",
				ShowAlert: true,
			})
		} else {
			newsettings := fn(callback, settings)
			botdb.SetChatSettings(chat, newsettings)

			b.Edit(callback.Message, generateSettingsMessageText(chat, newsettings), &tb.SendOptions{
				ParseMode:             tb.ModeMarkdown,
				ReplyMarkup:           generateSettingsReplyMarkup(chat, newsettings),
				DisableWebPagePreview: true,
			})

			b.Respond(callback, &tb.CallbackResponse{
				Text:      "Ok",
				ShowAlert: false,
			})
		}

	}
}

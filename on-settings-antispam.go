package main

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func sendAntispamSettingsMessage(messageToEdit *tb.Message, chatToSend *tb.Chat, chatToConfigure *tb.Chat, settings botdatabase.ChatSettings) {
	msg := generateAntispamSettingsMessageText(chatToConfigure, settings)
	var sendOpts = tb.SendOptions{
		ParseMode:             tb.ModeMarkdown,
		ReplyMarkup:           generateAntispamSettingsReplyMarkup(chatToConfigure, settings),
		DisableWebPagePreview: true,
	}
	if messageToEdit != nil {
		_, _ = b.Edit(messageToEdit, msg, &sendOpts)
	} else {
		_, _ = b.Send(chatToSend, msg, &sendOpts)
	}
}

func generateAntispamSettingsMessageText(chat *tb.Chat, settings botdatabase.ChatSettings) string {
	reply := strings.Builder{}

	reply.WriteString(fmt.Sprintf("Bot settings for chat %s (%d)\n\n", chat.Title, chat.ID))

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

	return reply.String()
}

func generateAntispamSettingsReplyMarkup(chat *tb.Chat, settings botdatabase.ChatSettings) *tb.ReplyMarkup {
	// TODO: move button creations in init function (eg. global buttons objects and handler)
	// Back settings
	backBtn := tb.InlineButton{
		Unique: "settings_back",
		Text:   "Back",
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&backBtn, func(callback *tb.Callback) {
		_ = b.Respond(callback)
		chatToConfigure, _ := b.ChatByID(callback.Data)

		// Back to main settings
		settings, _ := botdb.GetChatSetting(b, chatToConfigure)
		sendSettingsMessage(callback.Message, callback.Message.Chat, chatToConfigure, settings)
	})

	// On Join Chinese (TODO: add kick action)
	onJoinChineseKickButtonText := "✅ Ban Chinese on join"
	if settings.OnJoinChinese.Action != botdatabase.ACTION_NONE {
		onJoinChineseKickButtonText = "❌ Don't ban chinese joins"
	}
	onJoinChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_join",
		Text:   onJoinChineseKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onJoinChineseKickButton, CallbackAntispamSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
		if settings.OnJoinChinese.Action == botdatabase.ACTION_NONE {
			settings.OnJoinChinese = botdatabase.BotAction{
				Action: botdatabase.ACTION_BAN,
			}
		} else {
			settings.OnJoinChinese = botdatabase.BotAction{
				Action: botdatabase.ACTION_NONE,
			}
		}
		return settings
	}))

	// On Join Arabic (TODO: add kick action)
	onJoinArabicKickButtonText := "✅ Ban Arabic on join"
	if settings.OnJoinArabic.Action != botdatabase.ACTION_NONE {
		onJoinArabicKickButtonText = "❌ Don't ban arabs joins"
	}
	onJoinArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_join",
		Text:   onJoinArabicKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onJoinArabicKickButton, CallbackAntispamSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
		if settings.OnJoinArabic.Action == botdatabase.ACTION_NONE {
			settings.OnJoinArabic = botdatabase.BotAction{
				Action: botdatabase.ACTION_BAN,
			}
		} else {
			settings.OnJoinArabic = botdatabase.BotAction{
				Action: botdatabase.ACTION_NONE,
			}
		}
		return settings
	}))

	// On Message Chinese (TODO: add ban action)
	onMessageChineseKickButtonText := "✅ Kick Chinese msgs"
	if settings.OnMessageChinese.Action != botdatabase.ACTION_NONE {
		onMessageChineseKickButtonText = "❌ Don't kick chinese msgs"
	}
	onMessageChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_msgs",
		Text:   onMessageChineseKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onMessageChineseKickButton, CallbackAntispamSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
		if settings.OnMessageChinese.Action == botdatabase.ACTION_NONE {
			settings.OnMessageChinese = botdatabase.BotAction{
				Action: botdatabase.ACTION_KICK,
			}
		} else {
			settings.OnMessageChinese = botdatabase.BotAction{
				Action: botdatabase.ACTION_NONE,
			}
		}
		return settings
	}))

	// On Message Arabic (TODO: add ban action)
	onMessageArabicKickButtonText := "✅ Kick Arabic msgs"
	if settings.OnMessageArabic.Action != botdatabase.ACTION_NONE {
		onMessageArabicKickButtonText = "❌ Don't kick arabs msgs"
	}
	onMessageArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_msgs",
		Text:   onMessageArabicKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&onMessageArabicKickButton, CallbackAntispamSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
		if settings.OnMessageArabic.Action == botdatabase.ACTION_NONE {
			settings.OnMessageArabic = botdatabase.BotAction{
				Action: botdatabase.ACTION_KICK,
			}
		} else {
			settings.OnMessageArabic = botdatabase.BotAction{
				Action: botdatabase.ACTION_NONE,
			}
		}
		return settings
	}))

	// Enable CAS
	enableCASbuttonText := "❌ CAS disabled"
	if settings.OnBlacklistCAS.Action != botdatabase.ACTION_NONE {
		enableCASbuttonText = "✅ CAS enabled"
	}
	enableCASbutton := tb.InlineButton{
		Unique: "settings_enable_disable_cas",
		Text:   enableCASbuttonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	b.Handle(&enableCASbutton, CallbackAntispamSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
		if settings.OnBlacklistCAS.Action == botdatabase.ACTION_NONE {
			settings.OnBlacklistCAS = botdatabase.BotAction{
				Action: botdatabase.ACTION_KICK,
			}
		} else {
			settings.OnBlacklistCAS = botdatabase.BotAction{
				Action: botdatabase.ACTION_NONE,
			}
		}
		return settings
	}))

	return &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{onJoinChineseKickButton, onJoinArabicKickButton},
			{onMessageChineseKickButton, onMessageArabicKickButton},
			{enableCASbutton},
			{backBtn},
		},
	}
}

func CallbackAntispamSettings(fn func(*tb.Callback, botdatabase.ChatSettings) botdatabase.ChatSettings) func(callback *tb.Callback) {
	return func(callback *tb.Callback) {
		var err error
		chat := callback.Message.Chat
		if callback.Data != "" {
			chat, err = b.ChatByID(callback.Data)
			if err != nil {
				logger.WithError(err).Error("can't get chat by id")
				_ = b.Respond(callback, &tb.CallbackResponse{
					Text:      "Internal error",
					ShowAlert: true,
				})
				return
			}
		}

		settings, err := botdb.GetChatSetting(b, chat)
		if err != nil {
			logger.WithError(err).Error("Cannot get chat settings")
			_ = b.Respond(callback, &tb.CallbackResponse{
				Text:      "Internal error",
				ShowAlert: true,
			})
		} else if !settings.ChatAdmins.IsAdmin(callback.Sender) && !botdb.IsGlobalAdmin(callback.Sender) {
			logger.Warning("Non-admin is using a callback from the admin:", callback.Sender)
			_ = b.Respond(callback, &tb.CallbackResponse{
				Text:      "Not authorized",
				ShowAlert: true,
			})
		} else {
			newsettings := fn(callback, settings)
			_ = botdb.SetChatSettings(chat, newsettings)

			sendAntispamSettingsMessage(callback.Message, callback.Message.Chat, chat, newsettings)

			_ = b.Respond(callback, &tb.CallbackResponse{
				Text:      "Ok",
				ShowAlert: false,
			})
		}
	}
}

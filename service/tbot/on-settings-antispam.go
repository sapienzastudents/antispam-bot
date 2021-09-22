package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) sendAntispamSettingsMessage(messageToEdit *tb.Message, chatToSend *tb.Chat, chatToConfigure *tb.Chat, settings chatSettings) {
	msg := bot.generateAntispamSettingsMessageText(chatToConfigure, settings)
	var sendOpts = tb.SendOptions{
		ParseMode:             tb.ModeMarkdown,
		ReplyMarkup:           bot.generateAntispamSettingsReplyMarkup(chatToConfigure, settings),
		DisableWebPagePreview: true,
	}
	if messageToEdit != nil {
		_, _ = bot.telebot.Edit(messageToEdit, msg, &sendOpts)
	} else {
		_, _ = bot.telebot.Send(chatToSend, msg, &sendOpts)
	}
}

func (bot *telegramBot) generateAntispamSettingsMessageText(chat *tb.Chat, settings chatSettings) string {
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

func (bot *telegramBot) generateAntispamSettingsReplyMarkup(chat *tb.Chat, settings chatSettings) *tb.ReplyMarkup {
	// TODO: move button creations in init function (eg. global buttons objects and handler)
	// Back settings
	backBtn := tb.InlineButton{
		Unique: "settings_back",
		Text:   "Back",
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	bot.telebot.Handle(&backBtn, func(callback *tb.Callback) {
		_ = bot.telebot.Respond(callback)
		chatToConfigure, _ := bot.telebot.ChatByID(callback.Data)

		// Back to main settings
		settings, _ := bot.getChatSettings(chatToConfigure)
		bot.sendSettingsMessage(callback.Message, callback.Message.Chat, chatToConfigure, settings)
	})

	// On Join Chinese (TODO: add kick action)
	onJoinChineseKickButtonText := "✅ Ban Chinese on join"
	if settings.OnJoinChinese.Action != botdatabase.ActionNone {
		onJoinChineseKickButtonText = "❌ Don't ban chinese joins"
	}
	onJoinChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_join",
		Text:   onJoinChineseKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	bot.telebot.Handle(&onJoinChineseKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
		if settings.OnJoinChinese.Action == botdatabase.ActionNone {
			settings.OnJoinChinese = botdatabase.BotAction{
				Action: botdatabase.ActionBan,
			}
		} else {
			settings.OnJoinChinese = botdatabase.BotAction{
				Action: botdatabase.ActionNone,
			}
		}
		return settings
	}))

	// On Join Arabic (TODO: add kick action)
	onJoinArabicKickButtonText := "✅ Ban Arabic on join"
	if settings.OnJoinArabic.Action != botdatabase.ActionNone {
		onJoinArabicKickButtonText = "❌ Don't ban arabs joins"
	}
	onJoinArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_join",
		Text:   onJoinArabicKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	bot.telebot.Handle(&onJoinArabicKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
		if settings.OnJoinArabic.Action == botdatabase.ActionNone {
			settings.OnJoinArabic = botdatabase.BotAction{
				Action: botdatabase.ActionBan,
			}
		} else {
			settings.OnJoinArabic = botdatabase.BotAction{
				Action: botdatabase.ActionNone,
			}
		}
		return settings
	}))

	// On Message Chinese (TODO: add ban action)
	onMessageChineseKickButtonText := "✅ Kick Chinese msgs"
	if settings.OnMessageChinese.Action != botdatabase.ActionNone {
		onMessageChineseKickButtonText = "❌ Don't kick chinese msgs"
	}
	onMessageChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_msgs",
		Text:   onMessageChineseKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	bot.telebot.Handle(&onMessageChineseKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
		if settings.OnMessageChinese.Action == botdatabase.ActionNone {
			settings.OnMessageChinese = botdatabase.BotAction{
				Action: botdatabase.ActionKick,
			}
		} else {
			settings.OnMessageChinese = botdatabase.BotAction{
				Action: botdatabase.ActionNone,
			}
		}
		return settings
	}))

	// On Message Arabic (TODO: add ban action)
	onMessageArabicKickButtonText := "✅ Kick Arabic msgs"
	if settings.OnMessageArabic.Action != botdatabase.ActionNone {
		onMessageArabicKickButtonText = "❌ Don't kick arabs msgs"
	}
	onMessageArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_msgs",
		Text:   onMessageArabicKickButtonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	bot.telebot.Handle(&onMessageArabicKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
		if settings.OnMessageArabic.Action == botdatabase.ActionNone {
			settings.OnMessageArabic = botdatabase.BotAction{
				Action: botdatabase.ActionKick,
			}
		} else {
			settings.OnMessageArabic = botdatabase.BotAction{
				Action: botdatabase.ActionNone,
			}
		}
		return settings
	}))

	// Enable CAS
	enableCASbuttonText := "❌ CAS disabled"
	if settings.OnBlacklistCAS.Action != botdatabase.ActionNone {
		enableCASbuttonText = "✅ CAS enabled"
	}
	enableCASbutton := tb.InlineButton{
		Unique: "settings_enable_disable_cas",
		Text:   enableCASbuttonText,
		Data:   fmt.Sprintf("%d", chat.ID),
	}
	bot.telebot.Handle(&enableCASbutton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
		if settings.OnBlacklistCAS.Action == botdatabase.ActionNone {
			settings.OnBlacklistCAS = botdatabase.BotAction{
				Action: botdatabase.ActionKick,
			}
		} else {
			settings.OnBlacklistCAS = botdatabase.BotAction{
				Action: botdatabase.ActionNone,
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

func (bot *telegramBot) callbackAntispamSettings(fn func(*tb.Callback, chatSettings) chatSettings) func(callback *tb.Callback) {
	return func(callback *tb.Callback) {
		var err error
		chat := callback.Message.Chat
		if callback.Data != "" {
			chat, err = bot.telebot.ChatByID(callback.Data)
			if err != nil {
				bot.logger.WithError(err).Error("can't get chat by id")
				_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
					Text:      "Internal error",
					ShowAlert: true,
				})
				return
			}
		}

		settings, err := bot.getChatSettings(chat)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot get chat settings")
			_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
				Text:      "Internal error",
				ShowAlert: true,
			})
		} else if !settings.ChatAdmins.IsAdmin(callback.Sender) && !bot.db.IsGlobalAdmin(callback.Sender.ID) {
			bot.logger.Warning("Non-admin is using a callback from the admin:", callback.Sender)
			_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
				Text:      "Not authorized",
				ShowAlert: true,
			})
		} else {
			newsettings := fn(callback, settings)
			_ = bot.db.SetChatSettings(chat.ID, newsettings.ChatSettings)

			bot.sendAntispamSettingsMessage(callback.Message, callback.Message.Chat, chat, newsettings)

			_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
				Text:      "Ok",
				ShowAlert: false,
			})
		}
	}
}

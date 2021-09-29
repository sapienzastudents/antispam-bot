package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

// sendAntispamSettingsMessage sends the antispam settings message pane. This panel is an inner panel (i.e. you need to
// access settings first, then click on the antispam settings button).
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

// generateAntispamSettingsMessageText will generate the message text for sendAntispamSettingsMessage based on the
// current chat settings
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

// generateAntispamSettingsReplyMarkup will generate the button list for sendAntispamSettingsMessage based on the
// current chat settings. Buttons will change the bot settings for antispam
func (bot *telegramBot) generateAntispamSettingsReplyMarkup(chat *tb.Chat, settings chatSettings) *tb.ReplyMarkup {
	// TODO: move button creations in init function (eg. global buttons objects and handler)
	// Back settings
	backBtn := tb.InlineButton{
		Unique: "settings_back",
		Text:   "Back",
	}
	bot.handleAdminCallbackStateful(&backBtn, bot.backToSettingsFromCallback)

	// On Join Chinese (TODO: add kick action)
	onJoinChineseKickButtonText := "✅ Ban Chinese on join"
	if settings.OnJoinChinese.Action != botdatabase.ActionNone {
		onJoinChineseKickButtonText = "❌ Don't ban chinese joins"
	}
	onJoinChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_join",
		Text:   onJoinChineseKickButtonText,
		Data:   strconv.FormatInt(chat.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onJoinChineseKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
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
		Data:   strconv.FormatInt(chat.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onJoinArabicKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
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
		Data:   strconv.FormatInt(chat.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onMessageChineseKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
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
		Data:   strconv.FormatInt(chat.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onMessageArabicKickButton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
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
		Data:   strconv.FormatInt(chat.ID, 10),
	}
	bot.handleAdminCallbackStateful(&enableCASbutton, bot.callbackAntispamSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
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

// callbackAntispamSettings is a helper for callbacks in Antispam panel. It loads automatically the chat-to-edit settings, and
// save them at the end of the callback
func (bot *telegramBot) callbackAntispamSettings(fn func(*tb.Callback, chatSettings) chatSettings) func(*tb.Callback, State) {
	return func(callback *tb.Callback, state State) {
		settings, err := bot.getChatSettings(state.ChatToEdit)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot get chat settings")
			return
		}

		// Execute callback
		newsettings := fn(callback, settings)
		_ = bot.db.SetChatSettings(state.ChatToEdit.ID, newsettings.ChatSettings)
		_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
			Text:      "Ok",
			ShowAlert: false,
		})

		// Back to chat settings
		bot.sendAntispamSettingsMessage(callback.Message, callback.Message.Chat, state.ChatToEdit, newsettings)
	}
}

package bot

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/database"

	tb "gopkg.in/tucnak/telebot.v3"
)

// sendAntispamSettingsMessage sends the antispam settings message panel to the
// chat where the message is sent, localizing the text with the given language.
//
// This panel can be accessed when the user clicks on antispam settings button,
// insied the general settings panel.
func (bot *telegramBot) sendAntispamSettingsMessage(m *tb.Message, lang string, chatToConfigure *tb.Chat, settings chatSettings) {
	// Generate message text based on chatToConfigure.
	buf := strings.Builder{}

	buf.WriteString(fmt.Sprintf(bot.bundle.T(lang, "Bot settings for chat %s (%d)\n\n"), chatToConfigure.Title, chatToConfigure.ID))

	buf.WriteString("\n🇨🇳 " + bot.bundle.T(lang, "*Chinese* blocker:\n"))
	buf.WriteString(bot.bundle.T(lang, "On join: *"))
	buf.WriteString(prettyActionName(settings.OnJoinChinese, bot, lang))
	buf.WriteString("*\n")
	buf.WriteString(bot.bundle.T(lang, "On message: *"))
	buf.WriteString(prettyActionName(settings.OnMessageChinese, bot, lang))
	buf.WriteString("*\n")

	buf.WriteString("\n☪️  " + bot.bundle.T(lang, "*Arabic* blocker:\n"))
	buf.WriteString(bot.bundle.T(lang, "On join: *"))
	buf.WriteString(prettyActionName(settings.OnJoinArabic, bot, lang))
	buf.WriteString("*\n")
	buf.WriteString(bot.bundle.T(lang, "On message: *"))
	buf.WriteString(prettyActionName(settings.OnMessageArabic, bot, lang))
	buf.WriteString("*\n")

	buf.WriteString(bot.bundle.T(lang, "\nCAS-ban (see https://combot.org/cas/ ):\n"))
	buf.WriteString(bot.bundle.T(lang, "On any action: *"))
	buf.WriteString(prettyActionName(settings.OnBlacklistCAS, bot, lang))
	buf.WriteString("*\n")

	msg := buf.String()

	// Generate reply with button list based on current chat settings. Buttons
	// will change the bot settings for antispam.
	// Back settings
	backBtn := tb.InlineButton{
		Unique: "settings_back",
		Text:   "◀ " + bot.bundle.T(lang, "Back to settings"),
	}
	bot.handleAdminCallbackStateful(&backBtn, bot.backToSettingsFromCallback)

	// On Join Chinese (TODO: add kick action)
	onJoinChineseKickButtonText := "✅ " + bot.bundle.T(lang, "Ban Chinese on join")
	if settings.OnJoinChinese.Action != database.ActionNone {
		onJoinChineseKickButtonText = "❌ " + bot.bundle.T(lang, "Don't ban chinese joins")
	}
	onJoinChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_join",
		Text:   onJoinChineseKickButtonText,
		Data:   strconv.FormatInt(chatToConfigure.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onJoinChineseKickButton, bot.callbackAntispamSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
		if settings.OnJoinChinese.Action == database.ActionNone {
			settings.OnJoinChinese = database.BotAction{
				Action: database.ActionBan,
			}
		} else {
			settings.OnJoinChinese = database.BotAction{
				Action: database.ActionNone,
			}
		}
		return settings
	}))

	// On Join Arabic (TODO: add kick action)
	onJoinArabicKickButtonText := "✅ " + bot.bundle.T(lang, "Ban Arabic on join")
	if settings.OnJoinArabic.Action != database.ActionNone {
		onJoinArabicKickButtonText = "❌ " + bot.bundle.T(lang, "Don't ban arabs joins")
	}
	onJoinArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_join",
		Text:   onJoinArabicKickButtonText,
		Data:   strconv.FormatInt(chatToConfigure.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onJoinArabicKickButton, bot.callbackAntispamSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
		if settings.OnJoinArabic.Action == database.ActionNone {
			settings.OnJoinArabic = database.BotAction{
				Action: database.ActionBan,
			}
		} else {
			settings.OnJoinArabic = database.BotAction{
				Action: database.ActionNone,
			}
		}
		return settings
	}))

	// On Message Chinese (TODO: add ban action)
	onMessageChineseKickButtonText := "✅ " + bot.bundle.T(lang, "Kick Chinese msgs")
	if settings.OnMessageChinese.Action != database.ActionNone {
		onMessageChineseKickButtonText = "❌ " + bot.bundle.T(lang, "Don't kick chinese msgs")
	}
	onMessageChineseKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_chinese_on_msgs",
		Text:   onMessageChineseKickButtonText,
		Data:   strconv.FormatInt(chatToConfigure.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onMessageChineseKickButton, bot.callbackAntispamSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
		if settings.OnMessageChinese.Action == database.ActionNone {
			settings.OnMessageChinese = database.BotAction{
				Action: database.ActionKick,
			}
		} else {
			settings.OnMessageChinese = database.BotAction{
				Action: database.ActionNone,
			}
		}
		return settings
	}))

	// On Message Arabic (TODO: add ban action)
	onMessageArabicKickButtonText := "✅ " + bot.bundle.T(lang, "Kick Arabic msgs")
	if settings.OnMessageArabic.Action != database.ActionNone {
		onMessageArabicKickButtonText = "❌ " + bot.bundle.T(lang, "Don't kick arabs msgs")
	}
	onMessageArabicKickButton := tb.InlineButton{
		Unique: "settings_enable_disable_ban_arabic_on_msgs",
		Text:   onMessageArabicKickButtonText,
		Data:   strconv.FormatInt(chatToConfigure.ID, 10),
	}
	bot.handleAdminCallbackStateful(&onMessageArabicKickButton, bot.callbackAntispamSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
		if settings.OnMessageArabic.Action == database.ActionNone {
			settings.OnMessageArabic = database.BotAction{
				Action: database.ActionKick,
			}
		} else {
			settings.OnMessageArabic = database.BotAction{
				Action: database.ActionNone,
			}
		}
		return settings
	}))

	// Enable CAS
	enableCASbuttonText := "❌ " + bot.bundle.T(lang, "CAS disabled")
	if settings.OnBlacklistCAS.Action != database.ActionNone {
		enableCASbuttonText = "✅ " + bot.bundle.T(lang, "CAS enabled")
	}
	enableCASbutton := tb.InlineButton{
		Unique: "settings_enable_disable_cas",
		Text:   enableCASbuttonText,
		Data:   strconv.FormatInt(chatToConfigure.ID, 10),
	}
	bot.handleAdminCallbackStateful(&enableCASbutton, bot.callbackAntispamSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
		if settings.OnBlacklistCAS.Action == database.ActionNone {
			settings.OnBlacklistCAS = database.BotAction{
				Action: database.ActionKick,
			}
		} else {
			settings.OnBlacklistCAS = database.BotAction{
				Action: database.ActionNone,
			}
		}
		return settings
	}))

	reply := &tb.ReplyMarkup{
		InlineKeyboard: [][]tb.InlineButton{
			{onJoinChineseKickButton, onJoinArabicKickButton},
			{onMessageChineseKickButton, onMessageArabicKickButton},
			{enableCASbutton},
			{backBtn},
		},
	}

	// Send message.
	sendOpts := &tb.SendOptions{
		ParseMode:             tb.ModeMarkdown,
		ReplyMarkup:           reply,
		DisableWebPagePreview: true,
	}
	_, _ = bot.telebot.Edit(m, msg, sendOpts)
}

// prettyActionName returns an human-friendly name for the given action.
//
// If bot instance is nil it always returns English names, otherwhise returns
// the localized version for the given lang.
func prettyActionName(action database.BotAction, bot *telegramBot, lang string) string {
	T := func(s string) string { return s }
	if bot != nil {
		T = func(s string) string { return bot.bundle.T(lang, s) }
	}

	switch action.Action {
	case database.ActionMute:
		return "🔇 " + T("Mute")
	case database.ActionBan:
		return "🚷 " + T("Ban")
	case database.ActionDeleteMsg:
		return "✂️  " + T("Delete")
	case database.ActionKick:
		return "❗️ " + T("Kick")
	case database.ActionNone:
		return "💤 " + T("Do nothing")
	default:
		return "n/a"
	}
}

// callbackAntispamSettings returns a function suitable to be passed on
// handleAdminCallbackStateful.
//
// It is an helper function for callbacks in Antispam panel. Itloads
// automatically the chat-to-edit settings, and save them at the end of the
// callback.
func (bot *telegramBot) callbackAntispamSettings(fn func(tb.Context, chatSettings) chatSettings) func(tb.Context, State) {
	return func(ctx tb.Context, state State) {
		settings, err := bot.getChatSettings(state.ChatToEdit)
		if err != nil {
			bot.logger.WithError(err).Error("Cannot get chat settings")
			return
		}

		// Execute callback
		callback := ctx.Callback()
		newsettings := fn(ctx, settings)
		_ = bot.db.SetChatSettings(state.ChatToEdit.ID, newsettings.ChatSettings)
		_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
			Text:      "Ok",
			ShowAlert: false,
		})

		// Back to chat settings
		bot.sendAntispamSettingsMessage(callback.Message, callback.Sender.LanguageCode, state.ChatToEdit, newsettings)
	}
}

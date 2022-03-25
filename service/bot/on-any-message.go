package bot

import (
	"errors"
	"strconv"
	"strings"

	tb "gopkg.in/telebot.v3"
)

// onAnyMessage is triggered by any message in a group or in a private
// conversation.
func (bot *telegramBot) onAnyMessage(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}
	lang := ctx.Sender().LanguageCode

	// First, we need to retrieve the user state because we want to check
	// whether the user was previously in settings and he want to change the
	// group category or add a new bot admin. If this is the case, then this
	// message is the name for the new (sub) category or the ID of the new bot
	// admin.
	state := bot.getStateFor(m.Sender, m.Chat)
	if state.AddGlobalCategory || state.AddSubCategory {
		// Load chat settings for the chat that the user is editing (from
		// his state).
		settings, err := bot.getChatSettings(state.ChatToEdit)
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Warn("Failed to get chat settings")
			return
		}

		// Change category/subcategory for that chat.
		if state.AddGlobalCategory {
			categories := strings.Split(m.Text, "\n")
			settings.MainCategory = categories[0]
			if len(categories) > 1 {
				settings.SubCategory = categories[1]
			} else {
				settings.SubCategory = ""
			}
		} else {
			categories := strings.Split(m.Text, "\n")
			settings.SubCategory = categories[0]
		}

		// Save chat settings.
		err = bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Warn("Failed to save chat settings")
		}

		// Reset category naming flags in user state (we already changed the
		// (sub)category).
		state.AddSubCategory = false
		state.AddGlobalCategory = false
		state.Save()

		// Button for opening the settings menu again.
		settingsbt := tb.InlineButton{
			Unique: "back_to_settings",
			Text:   "◀ " + bot.bundle.T(lang, "Back to settings"),
		}
		bot.handleAdminCallbackStateful(&settingsbt, bot.backToSettingsFromCallback)

		_, _ = bot.telebot.Send(m.Chat, bot.bundle.T(lang, "Category saved"), &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{
				{settingsbt},
			},
		})
		return
	}
	if state.AddBotAdmin {
		_ = stateAddBotAdmin(bot, ctx, state)
		return
	}

	if !m.Private() { // On groups check message against antispam system.
		// G-Line check
		if banned, err := bot.db.IsUserBanned(m.Sender.ID); err == nil && banned {
			bot.banUser(m.Chat, m.Sender, settings, "user g-lined")
			bot.deleteMessage(m, settings, "user g-lined")
			return
		}

		// CAS ban check.
		if bot.cas != nil && bot.cas.IsBanned(m.Sender.ID) {
			bot.casDatabaseMatch.Inc()
			bot.performAction(m, m.Sender, settings, settings.OnBlacklistCAS, "CAS banned")
			return
		}

		// Check all text values against the antispam system.
		textvalues := []string{
			m.Text,
			m.Caption,
			m.Payload,
		}
		if m.Photo != nil {
			textvalues = append(textvalues, m.Photo.Caption)
		}
		if m.Audio != nil {
			textvalues = append(textvalues, m.Audio.Caption)
		}
		if m.Document != nil {
			textvalues = append(textvalues, m.Document.Caption)
		}
		if m.Video != nil {
			textvalues = append(textvalues, m.Video.Caption)
		}

		bot.spamFilter(m, settings, textvalues)
	}
}

func stateAddBotAdmin(bot *telegramBot, ctx tb.Context, state State) error {
	m := ctx.Message()
	lang := ctx.Sender().LanguageCode

	// This button is always shown also when the user send a invalid ID.
	var chatButtons [][]tb.InlineButton

	// Button to cancel the operation and go back to admin list panel.
	bt := tb.InlineButton{
		Unique: "add_admin_cancel",
		Text:   "❌ " + bot.bundle.T(lang, "Cancel"),
	}
	chatButtons = [][]tb.InlineButton{{bt}}
	bot.handleAdminCallbackStateful(&bt, func(ctx tb.Context, state State) {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return
		}
		state.AddBotAdmin = false
		state.Save()

		callback := ctx.Callback()
		bot.sendAdminsForSettings(callback.Sender, callback.Message)
	})

	// User can send an invalid ID.
	id, err := strconv.ParseInt(m.Text, 10, 64)
	if err != nil {
		msg := bot.bundle.T(lang, "The given user ID is not valid, please retry.")
		_, _ = bot.telebot.Send(m.Chat, msg, &tb.ReplyMarkup{InlineKeyboard: chatButtons})
		return nil
	}

	// User can send an ID that never contacted the bot, we can't get the user's
	// info.
	if _, err = bot.telebot.ChatByID(id); errors.Is(err, tb.ErrChatNotFound) {
		msg := bot.bundle.T(lang, "The given user ID never contacted the bot, please retry.")
		_, _ = bot.telebot.Send(m.Chat, msg, &tb.ReplyMarkup{InlineKeyboard: chatButtons})
		return nil
	}

	// Reset state flag. If there is an error on inserting ID, it is an
	// error on the server, in any case it's better to reset the state.
	state.AddBotAdmin = false
	state.Save()

	if err := bot.db.AddBotAdmin(id); err != nil {
		bot.logger.WithError(err).Error("Failed to add a new bot admin")
		msg := bot.bundle.T(lang, "Oops, I'm broken, please get in touch with my admin!")
		_, _ = bot.telebot.Send(m.Chat, msg, &tb.ReplyMarkup{InlineKeyboard: chatButtons})
		return nil
	}

	// Happy path: admin added without problems. There is a button to allow the
	// user to go back to admin list.
	bt = tb.InlineButton{
		Unique: "back_to_admins_settings",
		Text:   "◀ " + bot.bundle.T(lang, "Back to admin list"),
	}
	bot.handleAdminCallbackStateful(&bt, func(ctx tb.Context, state State) {
		if err := ctx.Respond(); err != nil {
			bot.logger.WithError(err).Error("Failed to respond to callback query")
			return
		}
		callback := ctx.Callback()
		bot.sendAdminsForSettings(callback.Sender, callback.Message)
	})
	chatButtons = [][]tb.InlineButton{{bt}}

	options := &tb.ReplyMarkup{InlineKeyboard: chatButtons}
	msg := bot.bundle.T(lang, "Admin added!")
	_, _ = bot.telebot.Send(m.Chat, msg, options)
	return nil
}

package tbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v3"
)

// onSettings is triggered by /settings command, sent either in a group or
// privately to the bot. The behaviour is different:
//
//	- When sent privately, onSettings replies with a list of groups where the
//	user is admin. Then, by clicking on one of these groups, the bot sends the
//	settings panel for that group.
//	- When sent in a public group, onSettings send the settings panel of that
//	group directly in the group itself. Note that in the latter case everyone is
//	able to see all settings (but only admins can change settings, of course).
func (bot *telegramBot) onSettings(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}

	bot.botCommandsRequestsTotal.WithLabelValues("settings").Inc()

	isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
		return
	}

	if !m.Private() && (isGlobalAdmin || settings.ChatAdmins.IsAdmin(m.Sender)) {
		// Messages in a chatroom: show settings panel for chatroom.
		bot.sendSettingsMessage(m.Sender, nil, m.Chat, m.Chat, settings)
	} else {
		// Private message: show group list when he is admin.
		bot.sendGroupListForSettings(m.Sender, nil, m.Chat, 0)
	}
}

// sendSettingsMessage sends the setting panel to the given chat. If
// messageToEdit is not nil, it edits that message instead of sending a new one.
//
// The message and button composition depends on current chat settings.
func (bot *telegramBot) sendSettingsMessage(user *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, chatToConfigure *tb.Chat, settings chatSettings) {
	var reply = strings.Builder{}
	var inlineKeyboard [][]tb.InlineButton

	// Create a new state for the user
	state := bot.newState(user, chatToSend)
	state.ChatToEdit = chatToConfigure
	state.Save()

	// Begin settings pane message
	reply.WriteString(fmt.Sprintf("Bot settings for chat %s (%d)\n\n", chatToConfigure.Title, chatToConfigure.ID))

	// Inform user about missing permissions
	me, err := bot.telebot.ChatMemberOf(chatToConfigure, bot.telebot.Me)
	if err != nil {
		bot.logger.WithError(err).WithFields(logrus.Fields{
			"chatid":   chatToConfigure.ID,
			"chattitle": chatToConfigure.Title,
		}).Error("Failed to get my info on chat")
		return
	}
	missingPrivileges := synthetizePrivileges(me)
	if me.Role != tb.Administrator {
		reply.WriteString("❌❌❌ The bot is not an admin! Admin permissions are needed for all functions\n")
	} else if len(missingPrivileges) != 0 {
		reply.WriteString("❌❌❌ Missing permissions:\n")
		for _, k := range missingPrivileges {
			reply.WriteString("• " + botPermissionsText[k] + "\n")
		}
		reply.WriteString("\n")
	}

	// Show settings only if the bot is an admin
	if me.Role == tb.Administrator {
		if me.CanDeleteMessages && me.CanRestrictMembers {
			// ============================== Enable / Disable bot button
			if settings.BotEnabled {
				reply.WriteString("✅ Bot enabled\n")
			} else {
				reply.WriteString("💤 Bot disabled\n")
			}
			enableDisableButtonText := "✅ Enable bot"
			if settings.BotEnabled {
				enableDisableButtonText = "❌ Disable bot"
			}
			enableDisableBotButton := tb.InlineButton{
				Unique: "settings_enable_disable_bot",
				Text:   enableDisableButtonText,
			}
			bot.handleAdminCallbackStateful(&enableDisableBotButton, bot.callbackSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
				settings.BotEnabled = !settings.BotEnabled
				return settings
			}))

			if settings.BotEnabled {
				// ============================== Go to antispam menu
				antispamSettingsButton := tb.InlineButton{
					Unique: "settings_goto_antispam",
					Text:   "✍️ Anti Spam",
				}
				bot.handleAdminCallbackStateful(&antispamSettingsButton, func(ctx tb.Context, state State) {
					callback := ctx.Callback()
					_ = bot.telebot.Respond(callback)

					settings, _ := bot.getChatSettings(state.ChatToEdit)
					bot.sendAntispamSettingsMessage(callback.Message, callback.Message.Chat, state.ChatToEdit, settings)
				})

				inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{enableDisableBotButton, antispamSettingsButton})
			} else {
				inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{enableDisableBotButton})
			}
		}

		if settings.BotEnabled && me.CanDeleteMessages {
			// ============================== Delete join messages
			if settings.OnJoinDelete {
				reply.WriteString("✅ Delete join message (after spam detection)\n")
			} else {
				reply.WriteString("💤 Do not delete join messages (after spam detection)\n")
			}
			deleteJoinMessagesText := "✅ Del join msgs"
			if settings.OnJoinDelete {
				deleteJoinMessagesText = "❌ Don't del join msgs"
			}
			deleteJoinMessages := tb.InlineButton{
				Unique: "settings_enable_disable_delete_on_join",
				Text:   deleteJoinMessagesText,
			}
			bot.handleAdminCallbackStateful(&deleteJoinMessages, bot.callbackSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
				settings.OnJoinDelete = !settings.OnJoinDelete
				return settings
			}))

			// ============================== Delete part messages
			if settings.OnLeaveDelete {
				reply.WriteString("✅ Delete leave message\n")
			} else {
				reply.WriteString("💤 Do not delete leave messages\n")
			}
			deleteLeaveMessagesText := "✅ Del leave msgs"
			if settings.OnLeaveDelete {
				deleteLeaveMessagesText = "❌ Don't del leave msgs"
			}
			deleteLeaveMessages := tb.InlineButton{
				Unique: "settings_enable_disable_delete_on_leave",
				Text:   deleteLeaveMessagesText,
			}
			bot.handleAdminCallbackStateful(&deleteLeaveMessages, bot.callbackSettings(func(context tb.Context, settings chatSettings) chatSettings {
				settings.OnLeaveDelete = !settings.OnLeaveDelete
				return settings
			}))

			inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{deleteJoinMessages, deleteLeaveMessages})
		}

		if me.CanInviteUsers {
			// ============================== Hide / show group in group lista
			if !settings.Hidden {
				reply.WriteString("\n👀 Group visible in group index")
			} else {
				reply.WriteString("\n⛔️ Group hidden from group index")
			}
			hideShowButtonText := "👀 Show group"
			if !settings.Hidden {
				hideShowButtonText = "⛔️ Hide group"
			}
			hideShowBotButton := tb.InlineButton{
				Unique: "settings_show_hide_group",
				Text:   hideShowButtonText,
			}
			bot.handleAdminCallbackStateful(&hideShowBotButton, bot.callbackSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
				settings.Hidden = !settings.Hidden
				return settings
			}))

			// ============================== Edit category
			reply.WriteString("\nGroup index category: ")
			if settings.MainCategory == "" {
				reply.WriteString("none\n")
			} else {
				reply.WriteString(settings.MainCategory)
				reply.WriteString(" ")
				reply.WriteString(settings.SubCategory)
				reply.WriteString("\n")
			}
			reply.WriteString("\n")
			editCategoryButtonText := "✏️ Edit category"
			editCategoryButton := tb.InlineButton{
				Unique: "settings_edit_group_category",
				Text:   editCategoryButtonText,
			}
			bot.handleAdminCallbackStateful(&editCategoryButton, bot.handleChangeCategory)

			inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{hideShowBotButton, editCategoryButton})
		}
	}

	reply.WriteString("\nLast updated: ")
	reply.WriteString(time.Now().Format(time.RFC1123Z))

	// ============================== Reload Group Info
	reloadGroupInfoBt := tb.InlineButton{
		Unique: "reload_group_info",
		Text:   "🛑 Restart bot",
	}
	bot.handleAdminCallbackStateful(&reloadGroupInfoBt, func(ctx tb.Context, state State) {
		_ = bot.DoCacheUpdateForChat(state.ChatToEdit.ID)

		callback := ctx.Callback()
		_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
			Text: "Bot restarted",
		})

		settings, _ := bot.getChatSettings(state.ChatToEdit)
		bot.sendSettingsMessage(user, callback.Message, state.chatWithTheUser, state.ChatToEdit, settings)
	})

	// ============================== Refresh
	settingsRefreshButton := tb.InlineButton{
		Unique: "settings_message_refresh",
		Text:   "🔄 Refresh",
	}
	bot.handleAdminCallbackStateful(&settingsRefreshButton, bot.callbackSettings(func(ctx tb.Context, settings chatSettings) chatSettings {
		return settings
	}))

	// ============================== Close settings
	closeBtn := tb.InlineButton{
		Unique: "settings_close",
		Text:   "🚪 Close",
	}
	bot.handleAdminCallbackStateful(&closeBtn, func(ctx tb.Context, state State) {
		_ = bot.telebot.Delete(ctx.Callback().Message)
	})

	inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{settingsRefreshButton, reloadGroupInfoBt})
	inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{closeBtn})

	sendOpts := tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		},
		DisableWebPagePreview: true,
	}
	if messageToEdit != nil {
		_, _ = bot.telebot.Edit(messageToEdit, reply.String(), &sendOpts)
	} else {
		_, _ = bot.telebot.Send(chatToSend, reply.String(), &sendOpts)
	}
}

// callbackSettings wraps the given function and returns a function suited to be
// passed in handleAdminCallbackStateful.
//
// It is an helper for callbacks in Settings panel, it loads automatically the
// ChatToEdit settings from the given State and save them at the end of the
// callback.
func (bot *telegramBot) callbackSettings(fn func(ctx tb.Context, settings chatSettings) chatSettings) func(tb.Context, State) {
	return func(ctx tb.Context, state State) {
		settings, err := bot.getChatSettings(state.ChatToEdit)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to get chat settings")
			return
		}

		// Execute callback
		newsettings := fn(ctx, settings)
		callback := ctx.Callback()
		_ = bot.db.SetChatSettings(state.ChatToEdit.ID, newsettings.ChatSettings)
		_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
			Text:      "Ok",
			ShowAlert: false,
		})

		// Back to chat settings
		bot.sendSettingsMessage(callback.Sender, callback.Message, callback.Message.Chat, state.ChatToEdit, newsettings)
	}
}

func (bot *telegramBot) backToSettingsFromCallback(ctx tb.Context, state State) {
	settings, _ := bot.getChatSettings(state.ChatToEdit)
	callback := ctx.Callback()
	bot.sendSettingsMessage(callback.Sender, callback.Message, callback.Message.Chat, state.ChatToEdit, settings)
}

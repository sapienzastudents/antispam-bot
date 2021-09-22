package tbot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	tb "gopkg.in/tucnak/telebot.v2"
)

const SettingsGroupListPageSize = 10

type inlineCategoryEdit struct {
	ChatID   int64
	Category string
}

func (bot *telegramBot) sendSettingsMessage(messageToEdit *tb.Message, chatToSend *tb.Chat, chatToConfigure *tb.Chat, settings chatSettings) {
	// TODO: move button creations in init function (eg. global buttons objects and handler)
	var reply = strings.Builder{}
	var inlineKeyboard [][]tb.InlineButton

	reply.WriteString(fmt.Sprintf("Bot settings for chat %s (%d)\n\n", chatToConfigure.Title, chatToConfigure.ID))

	// Check permissions
	me, _ := bot.telebot.ChatMemberOf(chatToConfigure, bot.telebot.Me)
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
				Data:   fmt.Sprintf("%d", chatToConfigure.ID),
			}
			bot.telebot.Handle(&enableDisableBotButton, bot.callbackSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
				settings.BotEnabled = !settings.BotEnabled
				return settings
			}))

			if settings.BotEnabled {
				// ============================== Go to antispam
				antispamSettingsButton := tb.InlineButton{
					Unique: "settings_goto_antispam",
					Text:   "✍️ Anti Spam",
					Data:   fmt.Sprintf("%d", chatToConfigure.ID),
				}
				bot.telebot.Handle(&antispamSettingsButton, func(callback *tb.Callback) {
					_ = bot.telebot.Respond(callback)
					chatToConfigure, _ := bot.telebot.ChatByID(callback.Data)
					settings, _ := bot.getChatSettings(chatToConfigure)

					bot.sendAntispamSettingsMessage(callback.Message, callback.Message.Chat, chatToConfigure, settings)
				})

				inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{enableDisableBotButton, antispamSettingsButton})
			} else {
				inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{enableDisableBotButton})
			}
		}

		if me.CanInviteUsers {
			// ============================== Hide / show group in group lista
			if !settings.Hidden {
				reply.WriteString("👀 Group visible\n")
			} else {
				reply.WriteString("⛔️ Group hidden\n")
			}
			hideShowButtonText := "👀 Show group"
			if !settings.Hidden {
				hideShowButtonText = "⛔️ Hide group"
			}
			hideShowBotButton := tb.InlineButton{
				Unique: "settings_show_hide_group",
				Text:   hideShowButtonText,
				Data:   fmt.Sprintf("%d", chatToConfigure.ID),
			}
			bot.telebot.Handle(&hideShowBotButton, bot.callbackSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
				settings.Hidden = !settings.Hidden
				return settings
			}))

			// ============================== Edit category
			reply.WriteString("\nCategory: ")
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
				Data:   fmt.Sprintf("%d", chatToConfigure.ID),
			}
			bot.telebot.Handle(&editCategoryButton, bot.handleChangeCategory)

			inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{hideShowBotButton, editCategoryButton})
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
				Data:   fmt.Sprintf("%d", chatToConfigure.ID),
			}
			bot.telebot.Handle(&deleteJoinMessages, bot.callbackSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
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
				Data:   fmt.Sprintf("%d", chatToConfigure.ID),
			}
			bot.telebot.Handle(&deleteLeaveMessages, bot.callbackSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
				settings.OnLeaveDelete = !settings.OnLeaveDelete
				return settings
			}))

			inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{deleteJoinMessages, deleteLeaveMessages})

		}
	}

	reply.WriteString("\nLast update: ")
	reply.WriteString(time.Now().Format(time.RFC1123Z))

	// ============================== Reload Group Info
	reloadGroupInfoBt := tb.InlineButton{
		Unique: "reload_group_info",
		Text:   "Reload group info",
		Data:   fmt.Sprintf("%d", chatToConfigure.ID),
	}
	bot.telebot.Handle(&reloadGroupInfoBt, func(callback *tb.Callback) {
		chatID, _ := strconv.Atoi(callback.Data)
		chatToConfigure, _ := bot.telebot.ChatByID(callback.Data)

		_ = bot.db.DoCacheUpdateForChat(bot.telebot, &tb.Chat{ID: int64(chatID)})

		_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
			Text: "Group information reloaded",
		})

		settings, _ := bot.getChatSettings(chatToConfigure)
		bot.sendSettingsMessage(callback.Message, chatToSend, chatToConfigure, settings)
	})

	// ============================== Refresh
	settingsRefreshButton := tb.InlineButton{
		Unique: "settings_message_refresh",
		Text:   "🔄 Refresh",
		Data:   fmt.Sprintf("%d", chatToConfigure.ID),
	}
	bot.telebot.Handle(&settingsRefreshButton, bot.callbackSettings(func(callback *tb.Callback, settings chatSettings) chatSettings {
		return settings
	}))

	// ============================== Close settings
	closeBtn := tb.InlineButton{
		Unique: "settings_close",
		Text:   "Close",
	}
	bot.telebot.Handle(&closeBtn, func(callback *tb.Callback) {
		_ = bot.telebot.Delete(callback.Message)
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

func (bot *telegramBot) onSettings(m *tb.Message, settings chatSettings) {
	bot.botCommandsRequestsTotal.WithLabelValues("settings").Inc()

	if !m.Private() && bot.db.IsGlobalAdmin(m.Sender.ID) || settings.ChatAdmins.IsAdmin(m.Sender) {
		// Messages in a chatroom
		bot.sendSettingsMessage(nil, m.Chat, m.Chat, settings)
	} else {
		// Private message
		bot.sendGroupListForSettings(m.Sender, nil, m.Chat, 0)
	}
}

func (bot *telegramBot) sendGroupListForSettings(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, page int) {
	var chatButtons [][]tb.InlineButton
	var showMore = false
	chatrooms, err := bot.db.ListMyChatrooms()
	if err != nil {
		bot.logger.WithError(err).Error("cant get chatroom list")
		return
	}

	// Sort chatrooms (to have a stable slice)
	sort.Slice(chatrooms, func(i, j int) bool {
		return chatrooms[i].Title < chatrooms[j].Title
	})

	isGlobalAdmin := bot.db.IsGlobalAdmin(sender.ID)

	// Pick chatrooms candidates (e.g. where the user has the admin permission)
	var candidates []*tb.Chat
	for _, x := range chatrooms {
		chatsettings, err := bot.getChatSettings(x)
		if err != nil {
			bot.logger.WithError(err).WithField("chat", x.ID).Warn("can't get chatroom settings")
			continue
		}
		if !isGlobalAdmin && !chatsettings.ChatAdmins.IsAdmin(sender) {
			continue
		}
		candidates = append(candidates, x)
	}

	// Slice the candidate list to the current page, if any
	if len(candidates) > (SettingsGroupListPageSize * (page + 1)) {
		candidates = candidates[SettingsGroupListPageSize*page : SettingsGroupListPageSize*(page+1)]
		showMore = true
	}
	if page > 0 && len(candidates) > SettingsGroupListPageSize*page {
		candidates = candidates[SettingsGroupListPageSize*page:]
	}

	// Create buttons
	for _, x := range candidates {
		btn := tb.InlineButton{
			Unique: fmt.Sprintf("select_chatid_%d", x.ID*-1),
			Text:   x.Title,
			Data:   fmt.Sprintf("%d", x.ID),
		}
		bot.telebot.Handle(&btn, func(callback *tb.Callback) {
			newchat, _ := bot.telebot.ChatByID(callback.Data)

			settings, _ := bot.getChatSettings(newchat)
			bot.sendSettingsMessage(callback.Message, callback.Message.Chat, newchat, settings)
		})
		chatButtons = append(chatButtons, []tb.InlineButton{btn})
	}

	var sendOptions = tb.SendOptions{}
	var msg string
	if len(chatButtons) == 0 {
		msg = "You are not an admin in a chat where the bot is."
	} else {
		if showMore {
			var bt = tb.InlineButton{
				Unique: "groups_settings_list_next",
				Text:   "Next ➡️",
				Data:   strconv.Itoa(page + 1),
			}
			chatButtons = append(chatButtons, []tb.InlineButton{bt})
			bot.telebot.Handle(&bt, func(callback *tb.Callback) {
				page, _ := strconv.Atoi(callback.Data)
				bot.sendGroupListForSettings(callback.Sender, callback.Message, callback.Message.Chat, page)
			})
		}
		if page >= 1 {
			var bt = tb.InlineButton{
				Unique: "groups_settings_list_prev",
				Text:   "⬅️ Prev",
				Data:   strconv.Itoa(page - 1),
			}
			chatButtons = append(chatButtons, []tb.InlineButton{bt})
			bot.telebot.Handle(&bt, func(callback *tb.Callback) {
				page, _ := strconv.Atoi(callback.Data)
				bot.sendGroupListForSettings(callback.Sender, callback.Message, callback.Message.Chat, page)
			})
		}

		var bt = tb.InlineButton{
			Unique: "groups_settings_list_close",
			Text:   "✖️ Close / Chiudi",
		}
		chatButtons = append(chatButtons, []tb.InlineButton{bt})
		bot.telebot.Handle(&bt, func(callback *tb.Callback) {
			_ = bot.telebot.Respond(callback)
			_ = bot.telebot.Delete(callback.Message)
		})

		msg = "Please select the chatroom:"
		sendOptions = tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: chatButtons,
			},
		}
	}

	if messageToEdit == nil {
		_, err = bot.telebot.Send(chatToSend, msg, &sendOptions)
	} else {
		_, err = bot.telebot.Edit(messageToEdit, msg, &sendOptions)
	}
	if err != nil {
		bot.logger.WithError(err).WithField("chatid", chatToSend.ID).Error("can't send/edit message for chat")
	}
}

func (bot *telegramBot) callbackSettings(fn func(*tb.Callback, chatSettings) chatSettings) func(callback *tb.Callback) {
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

			bot.sendSettingsMessage(callback.Message, callback.Message.Chat, chat, newsettings)

			_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
				Text:      "Ok",
				ShowAlert: false,
			})
		}

	}
}

func (bot *telegramBot) handleChangeCategory(callback *tb.Callback) {
	_ = bot.telebot.Respond(callback)
	chatID, _ := strconv.Atoi(callback.Data)

	categories, err := bot.db.GetChatTree(bot.telebot)
	if err != nil {
		bot.logger.WithError(err).WithField("chat", chatID).Error("can't get category tree")
		return
	}

	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_category",
		Data:   callback.Data,
	}
	bot.telebot.Handle(&customCategoryBt, func(callback *tb.Callback) {
		_ = bot.telebot.Respond(callback)
		chatID, _ := strconv.Atoi(callback.Data)
		_, _ = bot.telebot.Edit(callback.Message,
			"Scrivi il nome del corso di laurea.\n"+
				"Se vuoi inserire anche l'anno, mettilo in una seconda riga. Ad esempio:\n\n"+
				"Informatica (triennale)\n\noppure\n\nInformatica\nPrimo anno")

		bot.globaleditcat.Set(fmt.Sprint(callback.Sender.ID), inlineCategoryEdit{
			ChatID: int64(chatID),
		}, cache.DefaultExpiration)
	})

	buttons := [][]tb.InlineButton{{customCategoryBt}}
	for _, cat := range categories.GetSubCategoryList() {
		stateID := bot.newCallbackState(State{
			Chat:          &tb.Chat{ID: int64(chatID)}, // TODO: pass a chat type
			Category:      cat,
			SubCategories: categories.SubCategories[cat],
		})
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(cat),
			Data:   stateID,
		}

		bot.telebot.Handle(&bt, bot.CallbackStateful(func(callback *tb.Callback, state State) {
			bot.handleChangeSubCategory(callback, state)
		}))
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = bot.telebot.Edit(callback.Message, "Seleziona la categoria principale", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		bot.logger.WithError(err).Error("error sending message to the user in settings")
	}
}

func (bot *telegramBot) handleChangeSubCategory(callback *tb.Callback, state State) {
	_ = bot.telebot.Respond(callback)
	settings, _ := bot.getChatSettings(state.Chat)
	settings.MainCategory = state.Category
	err := bot.db.SetChatSettings(state.Chat.ID, settings.ChatSettings)
	if err != nil {
		bot.logger.WithError(err).WithField("chat", state.Chat.ID).Error("can't save chat settings")
		return
	}

	customCategorystate := bot.newCallbackState(State{
		Chat:     state.Chat,
		Category: state.Category,
	})
	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_subcategory",
		Data:   customCategorystate,
	}
	bot.telebot.Handle(&customCategoryBt, bot.CallbackStateful(func(callback *tb.Callback, state State) {
		_ = bot.telebot.Respond(callback)
		_, _ = bot.telebot.Edit(callback.Message, "Scrivi il nome della sotto-categoria.\n\nEsempio: Primo anno")

		bot.globaleditcat.Set(fmt.Sprint(callback.Sender.ID), inlineCategoryEdit{
			ChatID:   state.Chat.ID,
			Category: state.Category,
		}, cache.DefaultExpiration)
	}))

	noCategoryState := bot.newCallbackState(State{
		Chat:     state.Chat,
		Category: state.Category,
	})
	noCategoryBt := tb.InlineButton{
		Text:   "Nessuna sotto-categoria",
		Unique: "settings_no_sub_cat",
		Data:   noCategoryState,
	}
	bot.telebot.Handle(&noCategoryBt, bot.CallbackStateful(func(callback *tb.Callback, state State) {
		_ = bot.telebot.Respond(callback)
		settings, _ := bot.getChatSettings(state.Chat)
		settings.SubCategory = ""
		err := bot.db.SetChatSettings(state.Chat.ID, settings.ChatSettings)
		if err != nil {
			bot.logger.WithError(err).WithField("chat", state.Chat.ID).Error("can't save chat settings")
			return
		}

		settingsBt := tb.InlineButton{
			Text:   "Torna alle impostazioni",
			Unique: "back_to_settings",
			Data:   callback.Data,
		}
		bot.telebot.Handle(&settingsBt, bot.backToSettingsFromCallback)
		_, _ = bot.telebot.Edit(callback.Message, "Impostazioni salvate", &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
		})
	}))

	buttons := [][]tb.InlineButton{{customCategoryBt, noCategoryBt}}
	for _, cat := range state.SubCategories.GetSubCategoryList() {
		stateBt := bot.newCallbackState(State{
			Chat:        state.Chat,
			Category:    state.Category,
			SubCategory: cat,
		})
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(state.Category + cat),
			Data:   stateBt,
		}
		bot.telebot.Handle(&bt, bot.CallbackStateful(func(callback *tb.Callback, state State) {
			_ = bot.telebot.Respond(callback)
			settings, _ := bot.getChatSettings(state.Chat)
			settings.SubCategory = state.SubCategory
			err := bot.db.SetChatSettings(state.Chat.ID, settings.ChatSettings)
			if err != nil {
				bot.logger.WithError(err).WithField("chat", state.Chat.ID).Error("can't save chat settings")
				return
			}

			settingsBt := tb.InlineButton{
				Text:   "Torna alle impostazioni",
				Unique: "back_to_settings",
				Data:   callback.Data,
			}
			bot.telebot.Handle(&settingsBt, bot.backToSettingsFromCallback)
			_, _ = bot.telebot.Edit(callback.Message, "Impostazioni salvate", &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
			})
		}))
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = bot.telebot.Edit(callback.Message, "Seleziona la categoria interna", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		bot.logger.WithError(err).Error("error sending message to the user in settings")
	}
}

func (bot *telegramBot) backToSettingsFromCallback(callback *tb.Callback) {
	chat, err := bot.telebot.ChatByID(callback.Data)
	if err != nil {
		bot.logger.WithError(err).WithField("callback-data", callback.Data).Error("can't get chat information in backToSettingsFromCallback")
		return
	}
	settings, _ := bot.getChatSettings(chat)
	bot.sendSettingsMessage(callback.Message, callback.Message.Chat, chat, settings)
}

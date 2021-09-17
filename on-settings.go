package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

const SettingsGroupListPageSize = 10

type inlineCategoryEdit struct {
	ChatID   int64
	Category string
}

func sendSettingsMessage(messageToEdit *tb.Message, chatToSend *tb.Chat, chatToConfigure *tb.Chat, settings botdatabase.ChatSettings) {
	// TODO: move button creations in init function (eg. global buttons objects and handler)
	var reply = strings.Builder{}
	var inlineKeyboard [][]tb.InlineButton

	reply.WriteString(fmt.Sprintf("Bot settings for chat %s (%d)\n\n", chatToConfigure.Title, chatToConfigure.ID))

	// Check permissions
	me, _ := b.ChatMemberOf(chatToConfigure, b.Me)
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
			b.Handle(&enableDisableBotButton, callbackSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
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
				b.Handle(&antispamSettingsButton, func(callback *tb.Callback) {
					_ = b.Respond(callback)
					chatToConfigure, _ := b.ChatByID(callback.Data)
					settings, _ := botdb.GetChatSetting(b, chatToConfigure)

					sendAntispamSettingsMessage(callback.Message, callback.Message.Chat, chatToConfigure, settings)
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
			b.Handle(&hideShowBotButton, callbackSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
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
			b.Handle(&editCategoryButton, handleChangeCategory)

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
			b.Handle(&deleteJoinMessages, callbackSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
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
			b.Handle(&deleteLeaveMessages, callbackSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
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
	b.Handle(&reloadGroupInfoBt, func(callback *tb.Callback) {
		chatID, _ := strconv.Atoi(callback.Data)
		chatToConfigure, _ := b.ChatByID(callback.Data)

		_ = botdb.DoCacheUpdateForChat(b, &tb.Chat{ID: int64(chatID)})

		_ = b.Respond(callback, &tb.CallbackResponse{
			Text: "Group information reloaded",
		})

		settings, _ := botdb.GetChatSetting(b, chatToConfigure)
		sendSettingsMessage(callback.Message, chatToSend, chatToConfigure, settings)
	})

	// ============================== Refresh
	settingsRefreshButton := tb.InlineButton{
		Unique: "settings_message_refresh",
		Text:   "🔄 Refresh",
		Data:   fmt.Sprintf("%d", chatToConfigure.ID),
	}
	b.Handle(&settingsRefreshButton, callbackSettings(func(callback *tb.Callback, settings botdatabase.ChatSettings) botdatabase.ChatSettings {
		return settings
	}))

	// ============================== Close settings
	closeBtn := tb.InlineButton{
		Unique: "settings_close",
		Text:   "Close",
	}
	b.Handle(&closeBtn, func(callback *tb.Callback) {
		_ = b.Delete(callback.Message)
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
		_, _ = b.Edit(messageToEdit, reply.String(), &sendOpts)
	} else {
		_, _ = b.Send(chatToSend, reply.String(), &sendOpts)
	}
}

func onSettings(m *tb.Message, settings botdatabase.ChatSettings) {
	botCommandsRequestsTotal.WithLabelValues("settings").Inc()

	if !m.Private() && botdb.IsGlobalAdmin(m.Sender) || settings.ChatAdmins.IsAdmin(m.Sender) {
		// Messages in a chatroom
		sendSettingsMessage(nil, m.Chat, m.Chat, settings)
	} else {
		// Private message
		sendGroupListForSettings(m.Sender, nil, m.Chat, 0)
	}
}

func sendGroupListForSettings(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, page int) {
	var chatButtons [][]tb.InlineButton
	var showMore = false
	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("cant get chatroom list")
		return
	}

	// Sort chatrooms (to have a stable slice)
	sort.Slice(chatrooms, func(i, j int) bool {
		return chatrooms[i].Title < chatrooms[j].Title
	})

	isGlobalAdmin := botdb.IsGlobalAdmin(sender)

	// Pick chatrooms candidates (e.g. where the user has the admin permission)
	var candidates []*tb.Chat
	for _, x := range chatrooms {
		chatsettings, err := botdb.GetChatSetting(b, x)
		if err != nil {
			logger.WithError(err).WithField("chat", x.ID).Warn("can't get chatroom settings")
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
		b.Handle(&btn, func(callback *tb.Callback) {
			newchat, _ := b.ChatByID(callback.Data)

			settings, _ := botdb.GetChatSetting(b, newchat)
			sendSettingsMessage(callback.Message, callback.Message.Chat, newchat, settings)
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
			b.Handle(&bt, func(callback *tb.Callback) {
				page, _ := strconv.Atoi(callback.Data)
				sendGroupListForSettings(callback.Sender, callback.Message, callback.Message.Chat, page)
			})
		}
		if page >= 1 {
			var bt = tb.InlineButton{
				Unique: "groups_settings_list_prev",
				Text:   "⬅️ Prev",
				Data:   strconv.Itoa(page - 1),
			}
			chatButtons = append(chatButtons, []tb.InlineButton{bt})
			b.Handle(&bt, func(callback *tb.Callback) {
				page, _ := strconv.Atoi(callback.Data)
				sendGroupListForSettings(callback.Sender, callback.Message, callback.Message.Chat, page)
			})
		}

		var bt = tb.InlineButton{
			Unique: "groups_settings_list_close",
			Text:   "✖️ Close / Chiudi",
		}
		chatButtons = append(chatButtons, []tb.InlineButton{bt})
		b.Handle(&bt, func(callback *tb.Callback) {
			_ = b.Respond(callback)
			_ = b.Delete(callback.Message)
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
		_, err = b.Send(chatToSend, msg, &sendOptions)
	} else {
		_, err = b.Edit(messageToEdit, msg, &sendOptions)
	}
	if err != nil {
		logger.WithError(err).WithField("chatid", chatToSend.ID).Error("can't send/edit message for chat")
	}
}

func callbackSettings(fn func(*tb.Callback, botdatabase.ChatSettings) botdatabase.ChatSettings) func(callback *tb.Callback) {
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

			sendSettingsMessage(callback.Message, callback.Message.Chat, chat, newsettings)

			_ = b.Respond(callback, &tb.CallbackResponse{
				Text:      "Ok",
				ShowAlert: false,
			})
		}

	}
}

func handleChangeCategory(callback *tb.Callback) {
	_ = b.Respond(callback)
	chatID, _ := strconv.Atoi(callback.Data)

	categories, err := botdb.GetChatTree(b)
	if err != nil {
		logger.WithError(err).WithField("chat", chatID).Error("can't get category tree")
		return
	}

	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_category",
		Data:   callback.Data,
	}
	b.Handle(&customCategoryBt, func(callback *tb.Callback) {
		_ = b.Respond(callback)
		chatID, _ := strconv.Atoi(callback.Data)
		_, _ = b.Edit(callback.Message,
			"Scrivi il nome del corso di laurea.\n"+
				"Se vuoi inserire anche l'anno, mettilo in una seconda riga. Ad esempio:\n\n"+
				"Informatica (triennale)\n\noppure\n\nInformatica\nPrimo anno")

		globaleditcat.Set(fmt.Sprint(callback.Sender.ID), inlineCategoryEdit{
			ChatID: int64(chatID),
		}, cache.DefaultExpiration)
	})

	buttons := [][]tb.InlineButton{{customCategoryBt}}
	for _, cat := range categories.GetSubCategoryList() {
		stateID := newCallbackState(State{
			Chat:          &tb.Chat{ID: int64(chatID)}, // TODO: pass a chat type
			Category:      cat,
			SubCategories: categories.SubCategories[cat],
		})
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(cat),
			Data:   stateID,
		}

		b.Handle(&bt, CallbackStateful(func(callback *tb.Callback, state State) {
			handleChangeSubCategory(callback, state)
		}))
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = b.Edit(callback.Message, "Seleziona la categoria principale", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		logger.WithError(err).Error("error sending message to the user in settings")
	}
}

func handleChangeSubCategory(callback *tb.Callback, state State) {
	_ = b.Respond(callback)
	settings, _ := botdb.GetChatSetting(b, state.Chat)
	settings.MainCategory = state.Category
	err := botdb.SetChatSettings(state.Chat, settings)
	if err != nil {
		logger.WithError(err).WithField("chat", state.Chat.ID).Error("can't save chat settings")
		return
	}

	customCategorystate := newCallbackState(State{
		Chat:     state.Chat,
		Category: state.Category,
	})
	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_subcategory",
		Data:   customCategorystate,
	}
	b.Handle(&customCategoryBt, CallbackStateful(func(callback *tb.Callback, state State) {
		_ = b.Respond(callback)
		_, _ = b.Edit(callback.Message, "Scrivi il nome della sotto-categoria.\n\nEsempio: Primo anno")

		globaleditcat.Set(fmt.Sprint(callback.Sender.ID), inlineCategoryEdit{
			ChatID:   state.Chat.ID,
			Category: state.Category,
		}, cache.DefaultExpiration)
	}))

	noCategoryState := newCallbackState(State{
		Chat:     state.Chat,
		Category: state.Category,
	})
	noCategoryBt := tb.InlineButton{
		Text:   "Nessuna sotto-categoria",
		Unique: "settings_no_sub_cat",
		Data:   noCategoryState,
	}
	b.Handle(&noCategoryBt, CallbackStateful(func(callback *tb.Callback, state State) {
		_ = b.Respond(callback)
		settings, _ := botdb.GetChatSetting(b, state.Chat)
		settings.SubCategory = ""
		err := botdb.SetChatSettings(state.Chat, settings)
		if err != nil {
			logger.WithError(err).WithField("chat", state.Chat.ID).Error("can't save chat settings")
			return
		}

		settingsBt := tb.InlineButton{
			Text:   "Torna alle impostazioni",
			Unique: "back_to_settings",
			Data:   callback.Data,
		}
		b.Handle(&settingsBt, backToSettingsFromCallback)
		_, _ = b.Edit(callback.Message, "Impostazioni salvate", &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
		})
	}))

	buttons := [][]tb.InlineButton{{customCategoryBt, noCategoryBt}}
	for _, cat := range state.SubCategories.GetSubCategoryList() {
		stateBt := newCallbackState(State{
			Chat:        state.Chat,
			Category:    state.Category,
			SubCategory: cat,
		})
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(state.Category + cat),
			Data:   stateBt,
		}
		b.Handle(&bt, CallbackStateful(func(callback *tb.Callback, state State) {
			_ = b.Respond(callback)
			settings, _ := botdb.GetChatSetting(b, state.Chat)
			settings.SubCategory = state.SubCategory
			err := botdb.SetChatSettings(state.Chat, settings)
			if err != nil {
				logger.WithError(err).WithField("chat", state.Chat.ID).Error("can't save chat settings")
				return
			}

			settingsBt := tb.InlineButton{
				Text:   "Torna alle impostazioni",
				Unique: "back_to_settings",
				Data:   callback.Data,
			}
			b.Handle(&settingsBt, backToSettingsFromCallback)
			_, _ = b.Edit(callback.Message, "Impostazioni salvate", &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
			})
		}))
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = b.Edit(callback.Message, "Seleziona la categoria interna", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		logger.WithError(err).Error("error sending message to the user in settings")
	}
}

func backToSettingsFromCallback(callback *tb.Callback) {
	chat, err := b.ChatByID(callback.Data)
	if err != nil {
		logger.WithError(err).WithField("callback-data", callback.Data).Error("can't get chat information in backToSettingsFromCallback")
		return
	}
	settings, _ := botdb.GetChatSetting(b, chat)
	sendSettingsMessage(callback.Message, callback.Message.Chat, chat, settings)
}

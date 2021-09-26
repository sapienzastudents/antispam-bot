package tbot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// handleChangeCategory is the callback handler for "change category" button in settings pane
func (bot *telegramBot) handleChangeCategory(callback *tb.Callback, _ State) {
	_ = bot.telebot.Respond(callback)

	// Add "new category" button
	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_category",
	}
	bot.handleAdminCallbackStateful(&customCategoryBt, func(callback *tb.Callback, state State) {
		_ = bot.telebot.Respond(callback)
		_, _ = bot.telebot.Edit(callback.Message,
			"Scrivi il nome del corso di laurea.\n"+
				"Se vuoi inserire anche l'anno, mettilo in una seconda riga. Ad esempio:\n\n"+
				"Informatica (triennale)\n\noppure\n\nInformatica\nPrimo anno")

		state.AddGlobalCategory = true
		state.Save()
	})
	buttons := [][]tb.InlineButton{{customCategoryBt}}

	// Add existing categories
	categories, err := bot.db.GetChatTree(bot.telebot)
	if err != nil {
		bot.logger.WithError(err).Error("can't get category tree")
		return
	}

	for _, cat := range categories.GetSubCategoryList() {
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(cat),
			Data:   cat,
		}
		bot.handleAdminCallbackStateful(&bt, bot.handleChangeSubCategory)

		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = bot.telebot.Edit(callback.Message, "Seleziona la categoria principale", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		bot.logger.WithError(err).Error("error sending message to the user in settings")
	}
}

// handleChangeSubCategory is similar to handleChangeCategory, but for sub-categories
func (bot *telegramBot) handleChangeSubCategory(callback *tb.Callback, state State) {
	_ = bot.telebot.Respond(callback)
	settings, _ := bot.getChatSettings(state.ChatToEdit)
	settings.MainCategory = callback.Data
	err := bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
	if err != nil {
		bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("can't save chat settings")
		return
	}

	// Add new sub-category
	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_subcategory",
	}
	bot.handleAdminCallbackStateful(&customCategoryBt, func(callback *tb.Callback, state State) {
		_ = bot.telebot.Respond(callback)
		_, _ = bot.telebot.Edit(callback.Message, "Scrivi il nome della sotto-categoria.\n\nEsempio: Primo anno")

		state.AddSubCategory = true
		state.Save()
	})

	// No sub category button
	noCategoryBt := tb.InlineButton{
		Text:   "Nessuna sotto-categoria",
		Unique: "settings_no_sub_cat",
	}
	bot.handleAdminCallbackStateful(&noCategoryBt, func(callback *tb.Callback, state State) {
		_ = bot.telebot.Respond(callback)
		settings, _ := bot.getChatSettings(state.ChatToEdit)
		settings.SubCategory = ""
		err := bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("can't save chat settings")
			return
		}

		settingsBt := tb.InlineButton{
			Text:   "Torna alle impostazioni",
			Unique: "back_to_settings",
		}
		bot.handleAdminCallbackStateful(&settingsBt, bot.backToSettingsFromCallback)

		_, _ = bot.telebot.Edit(callback.Message, "Impostazioni salvate", &tb.ReplyMarkup{
			InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
		})
	})
	buttons := [][]tb.InlineButton{{customCategoryBt, noCategoryBt}}

	// Add sub-categories list
	rootChatTree, err := bot.db.GetChatTree(bot.telebot)
	if err != nil {
		bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("can't load chat tree")
		return
	}
	for cat := range rootChatTree.SubCategories[settings.MainCategory].SubCategories {
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(settings.MainCategory + cat),
			Data:   cat,
		}
		bot.handleAdminCallbackStateful(&bt, func(callback *tb.Callback, state State) {
			_ = bot.telebot.Respond(callback)
			settings, _ := bot.getChatSettings(state.ChatToEdit)
			settings.SubCategory = callback.Data
			err := bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
			if err != nil {
				bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("can't save chat settings")
				return
			}

			settingsBt := tb.InlineButton{
				Text:   "Torna alle impostazioni",
				Unique: "back_to_settings",
			}
			bot.handleAdminCallbackStateful(&settingsBt, bot.backToSettingsFromCallback)

			_, _ = bot.telebot.Edit(callback.Message, "Impostazioni salvate", &tb.ReplyMarkup{
				InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
			})
		})
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = bot.telebot.Edit(callback.Message, "Seleziona la categoria interna", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		bot.logger.WithError(err).Error("error sending message to the user in settings")
	}
}

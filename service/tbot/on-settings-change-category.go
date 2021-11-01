package tbot

import (
	tb "gopkg.in/tucnak/telebot.v3"
)

// handleChangeCategory is the callback handler for "change category" button in
// settings pane.
//
// It shows the list of current categories, and a button to create a new
// category. Each click on a category will trigger the list of subcategories
// handled by handleChangeSubCategory.
func (bot *telegramBot) handleChangeCategory(ctx tb.Context, state State) {
	callback := ctx.Callback()
	_ = bot.telebot.Respond(callback)

	// Add "new category" button
	customCategoryBt := tb.InlineButton{
		Text:   "✏️ Aggiungine una nuova",
		Unique: "settings_add_new_category",
	}
	bot.handleAdminCallbackStateful(&customCategoryBt, func(ctx tb.Context, state State) {
		callback := ctx.Callback()
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
	categories, err := bot.db.GetChatTree()
	if err != nil {
		bot.logger.WithError(err).Error("Failed to get category tree")
		return
	}

	for _, cat := range categories.GetSubCategoryList() {
		bt := tb.InlineButton{
			Text:   cat,
			Unique: sha1string(cat),
		}
		bot.handleAdminCallbackStateful(&bt, bot.handleChangeSubCategory(cat))

		buttons = append(buttons, []tb.InlineButton{bt})
	}

	_, err = bot.telebot.Edit(callback.Message, "Seleziona la categoria principale", &tb.ReplyMarkup{
		InlineKeyboard: buttons,
	})
	if err != nil {
		bot.logger.WithError(err).Error("Failed to message to the user in settings")
	}
}

// handleChangeSubCategory is lke handleChangeCategory, but for sub-categories.
func (bot *telegramBot) handleChangeSubCategory(categoryName string) func(ctx tb.Context, state State) {
	return func(ctx tb.Context, state State) {
		callback := ctx.Callback()
		_ = bot.telebot.Respond(callback)

		settings, _ := bot.getChatSettings(state.ChatToEdit)
		settings.MainCategory = categoryName
		err := bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("Failed to save chat settings")
			return
		}

		// Add new sub-category
		customCategoryBt := tb.InlineButton{
			Text:   "✏️ Aggiungine una nuova",
			Unique: "settings_add_new_subcategory",
		}
		bot.handleAdminCallbackStateful(&customCategoryBt, func(ctx tb.Context, state State) {
			callback := ctx.Callback()
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
		bot.handleAdminCallbackStateful(&noCategoryBt, func(ctx tb.Context, state State) {
			callback := ctx.Callback()
			_ = bot.telebot.Respond(callback)
			settings, _ := bot.getChatSettings(state.ChatToEdit)
			settings.SubCategory = ""
			err := bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
			if err != nil {
				bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("Failed to save chat settings")
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
		rootChatTree, err := bot.db.GetChatTree()
		if err != nil {
			bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("Failed to load chat tree")
			return
		}
		for cat := range rootChatTree.SubCategories[settings.MainCategory].SubCategories {
			bt := tb.InlineButton{
				Text:   cat,
				Unique: sha1string(settings.MainCategory + cat),
			}
			bot.handleAdminCallbackStateful(&bt, func(subCategoryName string) func(ctx tb.Context, state State) {
				return func(ctx tb.Context, state State) {
					callback := ctx.Callback()
					_ = bot.telebot.Respond(callback)

					settings, _ := bot.getChatSettings(state.ChatToEdit)
					settings.SubCategory = subCategoryName
					err := bot.db.SetChatSettings(state.ChatToEdit.ID, settings.ChatSettings)
					if err != nil {
						bot.logger.WithError(err).WithField("chatid", state.ChatToEdit.ID).Error("Failed to save chat settings")
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
				}
			}(cat))
			buttons = append(buttons, []tb.InlineButton{bt})
		}

		_, err = bot.telebot.Edit(callback.Message, "Seleziona la categoria interna", &tb.ReplyMarkup{
			InlineKeyboard: buttons,
		})
		if err != nil {
			bot.logger.WithError(err).Error("Failed to send message to the user in settings")
		}
	}
}

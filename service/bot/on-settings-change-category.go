package bot

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

	lang := ctx.Sender().LanguageCode

	// Add "new category" button
	customCategoryBt := tb.InlineButton{
		Text:   "✏️  " + bot.bundle.T(lang, "Add new category"),
		Unique: "settings_add_new_category",
	}
	bot.handleAdminCallbackStateful(&customCategoryBt, func(ctx tb.Context, state State) {
		callback := ctx.Callback()
		_ = bot.telebot.Respond(callback)

		lang := ctx.Sender().LanguageCode
		msg := bot.bundle.T(lang, "Write the degree course name. You can also write the year, but write it in a second line. As example:\n\nComputer Science (bachelor)\n\nOr\n\n Computer Science\nFirst Year")

		_, _ = bot.telebot.Edit(callback.Message, msg)
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

	_, err = bot.telebot.Edit(callback.Message,
		bot.bundle.T(lang, "Select main category"),
		&tb.ReplyMarkup{InlineKeyboard: buttons})
	if err != nil {
		bot.logger.WithError(err).Error("Failed to message to the user in settings")
	}
}

// handleChangeSubCategory is like handleChangeCategory, but for sub-categories.
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

		lang := ctx.Sender().LanguageCode

		// Add new sub-category
		customCategoryBt := tb.InlineButton{
			Text:   "✏️  " + bot.bundle.T(lang, "Add new subcategory"),
			Unique: "settings_add_new_subcategory",
		}
		bot.handleAdminCallbackStateful(&customCategoryBt, func(ctx tb.Context, state State) {
			callback := ctx.Callback()
			lang := ctx.Sender().LanguageCode
			_ = bot.telebot.Respond(callback)

			msg := bot.bundle.T(lang, "Write the subcategory name. As example:\n\nFirst year")
			_, _ = bot.telebot.Edit(callback.Message, msg)

			state.AddSubCategory = true
			state.Save()
		})

		// No sub category button
		noCategoryBt := tb.InlineButton{
			Text:   bot.bundle.T(lang, "No subcategory"),
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

			lang := ctx.Sender().LanguageCode
			settingsBt := tb.InlineButton{
				Text:   "◀ " + bot.bundle.T(lang, "Back to settings"),
				Unique: "back_to_settings",
			}
			bot.handleAdminCallbackStateful(&settingsBt, bot.backToSettingsFromCallback)

			msg := bot.bundle.T(lang, "Settings saved")
			_, _ = bot.telebot.Edit(callback.Message, msg, &tb.ReplyMarkup{
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

					lang := ctx.Sender().LanguageCode
					settingsBt := tb.InlineButton{
						Text:   "◀ " + bot.bundle.T(lang, "Back to settings"),
						Unique: "back_to_settings",
					}
					bot.handleAdminCallbackStateful(&settingsBt, bot.backToSettingsFromCallback)

					msg := bot.bundle.T(lang, "Settings saved")
					_, _ = bot.telebot.Edit(callback.Message, msg, &tb.ReplyMarkup{
						InlineKeyboard: [][]tb.InlineButton{{settingsBt}},
					})
				}
			}(cat))
			buttons = append(buttons, []tb.InlineButton{bt})
		}

		msg := bot.bundle.T(lang, "Select subcategory")
		_, err = bot.telebot.Edit(callback.Message, msg, &tb.ReplyMarkup{
			InlineKeyboard: buttons,
		})
		if err != nil {
			bot.logger.WithError(err).Error("Failed to send message to the user in settings")
		}
	}
}

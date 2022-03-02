package bot

import tb "gopkg.in/telebot.v3"

// AddBlacklist adds the chat the user is editing (state.ChatToEdit) to the
// blacklist.
//
// Only bot admins can do this action.
func (bot *telegramBot) AddBlacklist(ctx tb.Context, state State) {
	sender := ctx.Sender()
	logger := bot.logger.WithField("user_id", sender.ID)
	lang := sender.LanguageCode

	// Only bot admins can blacklist a group.
	if is, err := bot.db.IsBotAdmin(sender.ID); err != nil {
		logger.WithError(err).Error("Failed to check if the user is a bot admin")
		err := ctx.Respond(&tb.CallbackResponse{
			Text:      bot.bundle.T(lang, "Failed to check if you are a bot admin"),
			ShowAlert: true,
		})
		if err != nil {
			bot.logger.WithError(err).Error("Failed to reply to a callback")
		}
		return
	} else if !is {
		err := ctx.Respond(&tb.CallbackResponse{
			Text: bot.bundle.T(lang, "Only bot admins can blacklist a group!"),
		})
		if err != nil {
			bot.logger.WithError(err).Error("Failed to reply to a callback")
		}
		return
	}

	blacklisted := state.ChatToEdit
	logger = logger.WithField("chat_id", blacklisted.ID)
	if err := bot.db.AddBlacklist(blacklisted); err != nil {
		logger.WithError(err).Error("Failed to add group to blacklist")
		err := ctx.Respond(&tb.CallbackResponse{
			Text: bot.bundle.T(lang, "Failed to add group to the blacklist!"),
		})
		if err != nil {
			bot.logger.WithError(err).Error("Failed to reply to a callback")
		}
	}

	err := ctx.Respond(&tb.CallbackResponse{
		Text: bot.bundle.T(lang, "Group added to the blacklist!"),
	})
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply to a callback")
	}

	// Close admin panel, because the info are now done.
	logger.Info("Group added to the blacklist")
	_ = bot.telebot.Delete(ctx.Callback().Message)
}

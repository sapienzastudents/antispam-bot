package bot

import (
	"fmt"
	"sort"
	"strconv"

	tb "gopkg.in/telebot.v3"
)

const BlacklistPageSize = 10

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

// SendBlacklist sends a message with a list of groups that are blacklisted.
//
// It works only on private chats.The list is a flat list of groups. After
// clicking on a button of the list, a confirmation panel to remove the group
// from the blacklist will be sent to him.
//
// If messageToEdit is nil, it will send the list to the sender's private chat.
func (bot *telegramBot) sendBlacklist(sender *tb.User, messageToEdit *tb.Message, page int) {
	// Only bot admins can see the blacklist.
	if is, err := bot.db.IsBotAdmin(sender.ID); err != nil {
		bot.logger.WithError(err).Error("Failed to check if the user is a bot admin")
		return
	} else if !is {
		bot.logger.Warn("This user triggered the blacklist but it is not a bot admin!")
		return
	}

	blacklist, err := bot.db.ListBlacklist()
	if err != nil {
		bot.logger.WithError(err).Error("Failed to get blacklist")
	}

	// The blacklist can be huge and a message have a limit for the number of
	// buttons, so we need to paging it.

	// Sort blacklist to have a stable slice.
	sort.Slice(blacklist, func(i, j int) bool {
		return blacklist[i].Title < blacklist[j].Title
	})

	// Slice the blacklist to the max number of pages, starting from page.
	showMore := false
	if len(blacklist) > (SettingsGroupListPageSize * (page + 1)) {
		blacklist = blacklist[SettingsGroupListPageSize*page : SettingsGroupListPageSize*(page+1)]
		showMore = true
	}
	if page > 0 && len(blacklist) > SettingsGroupListPageSize*page {
		blacklist = blacklist[SettingsGroupListPageSize*page:]
	}

	// Create buttons.
	var chatButtons [][]tb.InlineButton
	for _, x := range blacklist {
		btn := tb.InlineButton{
			Unique: fmt.Sprintf("select_blacklist_chatid_%d", x.ID*-1),
			Text:   x.Title,
			Data:   strconv.FormatInt(x.ID, 10),
		}
		bot.telebot.Handle(&btn, func(ctx tb.Context) error {
			callback := ctx.Callback()

			if err := ctx.Respond(); err != nil {
				bot.logger.WithError(err).Error("Failed to respond to callback query")
				return err
			}

			// Parse chat's id from callback data.
			id, err := strconv.ParseInt(callback.Data, 10, 64)
			if err != nil {
				bot.logger.WithError(err).Error("Failed to parse callback data")
				return err
			}

			bot.sendBlacklistRemoval(callback.Sender, callback.Message, id)
			return nil
		})
		chatButtons = append(chatButtons, []tb.InlineButton{btn})
	}

	lang := sender.LanguageCode

	sendOptions := &tb.SendOptions{}
	msg := ""

	if len(chatButtons) == 0 { // Special case: blacklist empty.
		bt := tb.InlineButton{
			Unique: "blacklist_close",
			Text:   "🚪 " + bot.bundle.T(lang, "Close"),
		}
		chatButtons = append(chatButtons, []tb.InlineButton{bt})
		bot.telebot.Handle(&bt, func(ctx tb.Context) error {
			callback := ctx.Callback()
			_ = bot.telebot.Respond(callback)
			_ = bot.telebot.Delete(callback.Message)
			return nil
		})

		msg = bot.bundle.T(lang, "Blacklist is empty.")
		sendOptions = &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: chatButtons,
			},
		}
	} else {
		if page >= 1 {
			bt := tb.InlineButton{
				Unique: "blacklist_prev",
				Text:   "⬅️  " + bot.bundle.T(lang, "Prev"),
				Data:   strconv.Itoa(page - 1),
			}
			chatButtons = append(chatButtons, []tb.InlineButton{bt})
			bot.telebot.Handle(&bt, func(ctx tb.Context) error {
				callback := ctx.Callback()
				page, err := strconv.Atoi(callback.Data)
				if err != nil {
					return err
				}
				bot.sendBlacklist(callback.Sender, callback.Message, page)
				return nil
			})
		}
		if showMore {
			bt := tb.InlineButton{
				Unique: "blaklist_next",
				Text:   bot.bundle.T(lang, "Next") + " ➡️",
				Data:   strconv.Itoa(page + 1),
			}
			chatButtons = append(chatButtons, []tb.InlineButton{bt})
			bot.telebot.Handle(&bt, func(ctx tb.Context) error {
				callback := ctx.Callback()
				page, err := strconv.Atoi(callback.Data)
				if err != nil {
					return err
				}
				bot.sendBlacklist(callback.Sender, callback.Message, page)
				return nil
			})
		}

		bt := tb.InlineButton{
			Unique: "blacklist_close",
			Text:   "🚪 " + bot.bundle.T(lang, "Close"),
		}
		chatButtons = append(chatButtons, []tb.InlineButton{bt})
		bot.telebot.Handle(&bt, func(ctx tb.Context) error {
			callback := ctx.Callback()
			_ = bot.telebot.Respond(callback)
			_ = bot.telebot.Delete(callback.Message)
			return nil
		})

		msg = bot.bundle.T(lang, "Please select the group you want to remove from the blacklist:")
		sendOptions = &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: chatButtons,
			},
		}
	}

	if messageToEdit == nil {
		if _, err = bot.telebot.Send(sender, msg, sendOptions); err != nil {
			bot.logger.WithError(err).Error("Failed to send blacklist message")
		}
	} else {
		if _, err = bot.telebot.Edit(messageToEdit, msg, sendOptions); err != nil {
			bot.logger.WithError(err).Error("Failed to edit blacklist message")
		}
	}
}

// sendBlacklistRemoval sends a confirmation message to remove the given chat's
// id from the blacklist.
//
// The confirmation message is sent editing messageToEdit. After clicking a
// button, the blacklist list will be sent.
func (bot *telegramBot) sendBlacklistRemoval(sender *tb.User, message *tb.Message, id int64) {
	lang := sender.LanguageCode

	// Only bot admins can remove a group from a blacklist.
	if is, err := bot.db.IsBotAdmin(sender.ID); err != nil {
		bot.logger.WithError(err).Error("Failed to check if the user is a bot admin")
		return
	} else if !is {
		bot.logger.Warn("This user triggered the blacklist but it is not a bot admin!")
		return
	}

	var chatButtons [][]tb.InlineButton

	// "Yes" (want to remove the group from the blacklist) button.
	yesBt := tb.InlineButton{
		Unique: "confirm_blacklist_yes",
		Text:   "✅ " + bot.bundle.T(lang, "Yes"),
		Data:   strconv.FormatInt(id, 10),
	}
	bot.telebot.Handle(&yesBt, func(ctx tb.Context) error {
		callback := ctx.Callback()

		id, err := strconv.ParseInt(callback.Data, 10, 64)
		if err != nil {
			bot.logger.WithError(err).Error("Failed to parse callback data as int64")
			ctx.Respond(&tb.CallbackResponse{
				Text: "Internal server error, contact an admin!",
			})
			return err
		}

		if err := bot.db.DeleteBlacklist(id); err != nil {
			bot.logger.WithError(err).WithField("chat_id", id).Error("Failed to remove group from blacklist")
			ctx.Respond(&tb.CallbackResponse{
				Text: "Internal server error, contact an admin!",
			})
			return err
		}

		ctx.Respond(&tb.CallbackResponse{
			Text: "Chat removed from the blacklist!",
		})
		bot.sendBlacklist(callback.Sender, callback.Message, 0)
		return nil
	})

	// "No" (do not want to remove the group from the blacklist) button.
	noBt := tb.InlineButton{
		Unique: "confirm_blacklist_no",
		Text:   "❌ " + bot.bundle.T(lang, "No"),
	}
	bot.telebot.Handle(&noBt, func(ctx tb.Context) error {
		callback := ctx.Callback()
		ctx.Respond(&tb.CallbackResponse{
			Text: "Chat NOT removed from the blacklist!",
		})
		bot.sendBlacklist(callback.Sender, callback.Message, 0)
		return nil
	})

	chatButtons = append(chatButtons, []tb.InlineButton{yesBt, noBt})

	// We need to get chat's info to retrieve the title.
	chat, err := bot.db.GetBlacklist(id)
	if err != nil {
		bot.logger.WithField("blacklist_id", id).WithError(err).Error("Failed to get blacklisted chat")
		return
	}

	rawMsg := bot.bundle.T(lang, "Do you want to remove %q group from the blacklist?")
	msg := fmt.Sprintf(rawMsg, chat.Title)
	options := &tb.ReplyMarkup{
		InlineKeyboard: chatButtons,
	}
	if _, err := bot.telebot.Edit(message, msg, options); err != nil {
		bot.logger.WithError(err).Error("Failed to edit blacklist message to confirmation message")
	}
}

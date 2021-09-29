package tbot

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"sort"
	"strconv"
)

const SettingsGroupListPageSize = 10

// sendGroupListForSettings creates the group list for settings page, which means that this list is a flat list (no
// categories involved) which contains only groups where the user is allowed to configure things (e.g. groups where
// he/she is an admin). It's sent in private to the user. After clicking on a button of the list (i.e. on a group), the
// settings pane will be sent to him/her
func (bot *telegramBot) sendGroupListForSettings(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, page int) {

	// Rationale: we need to list all groups where the user is admin. This list can be _huge_. So we need to paging it.
	// In order to page the list of groups, we need to filter the list before paging. So this function will:
	// 1. Get the full list of chatrooms of the bot
	// 2. Check if the user is an admin for each chatroom (creating a list of "candidates" chats)
	// 3. Slice the list to the max number of pages (in constant SettingsGroupListPageSize), starting from the page
	//    indicated
	// 4. Then, create the message (text + list of buttons)

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

	isGlobalAdmin, err := bot.db.IsGlobalAdmin(sender.ID)
	if err != nil {
		bot.logger.WithError(err).Error("can't check if the user is a global admin")
		return
	}

	// Pick chatrooms candidates (e.g. where the user has the admin permission)
	var candidates []*tb.Chat
	for _, x := range chatrooms {
		if !isGlobalAdmin {
			chatsettings, err := bot.getChatSettings(x)
			if err != nil {
				bot.logger.WithError(err).WithField("chat", x.ID).Warn("can't get chatroom settings")
				continue
			}
			if !chatsettings.ChatAdmins.IsAdmin(sender) {
				continue
			}
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
			Data:   strconv.FormatInt(x.ID, 10),
		}
		bot.telebot.Handle(&btn, func(callback *tb.Callback) {
			newchat, _ := bot.telebot.ChatByID(callback.Data)

			settings, _ := bot.getChatSettings(newchat)
			bot.sendSettingsMessage(callback.Sender, callback.Message, callback.Message.Chat, newchat, settings)
		})
		chatButtons = append(chatButtons, []tb.InlineButton{btn})
	}

	var sendOptions = tb.SendOptions{}
	var msg string
	if len(chatButtons) == 0 {
		msg = "You are not an admin in a chat where the bot is."
	} else {
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

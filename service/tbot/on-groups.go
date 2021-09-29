package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// onGroups fires when the user sends a /groups command in private or in a group. See sendGroupListForLinks as this
// function is an alias for sendGroupListForLinks
func (bot *telegramBot) onGroups(m *tb.Message, _ chatSettings) {
	bot.sendGroupListForLinks(m.Sender, nil, m.Chat, m)
}

// sendGroupListForLinks sends a list of categories as buttons. When clicking on a category/button, the message is
// replaced with the list of groups, divided in subcategories
func (bot *telegramBot) sendGroupListForLinks(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, messageFromUser *tb.Message) {
	bot.botCommandsRequestsTotal.WithLabelValues("groups").Inc()

	categoryTree, err := bot.db.GetChatTree()
	if err != nil {
		bot.logger.WithError(err).Error("Error getting chatroom list")
		msg, _ := bot.telebot.Send(chatToSend, "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-(")
		bot.setMessageExpiry(msg, 30*time.Second)
		return
	}

	// Rationale: in Telegram, we can add a matrix of InlineButtons. Each row (outer index) can be composed by multiple
	// buttons (inner index). However, categories can be very long, so we stick with one button per row.
	//
	// Rationale: we support only two layers of "categories": now we draw the first one as a list of buttons. Note that
	// we need to show the right category for the right button, so:
	// * we register the button for the category using the sha1 of the name (the whole name might be too long, or
	//   contains illegal chars)
	// * we register a callback handler using a closure to bind the category variable (so we can show the right
	//   subcategory list)
	//
	// Don't try to use the custom "Data" field for buttons here: it doesn't work due some limitations on Telegram side.

	var buttons [][]tb.InlineButton
	for _, category := range categoryTree.GetSubCategoryList() {
		var bt = tb.InlineButton{
			Unique: sha1string(category),
			Text:   category,
		}
		buttons = append(buttons, []tb.InlineButton{bt})

		bot.telebot.Handle(&bt, func(cat botdatabase.ChatCategoryTree) func(callback *tb.Callback) {
			return func(callback *tb.Callback) {
				bot.showCategory(callback.Message, cat, false)
				_ = bot.telebot.Respond(callback)
			}
		}(categoryTree.SubCategories[category]))
	}

	// Global admins are able to see a special category which contains all groups without a category. This is for
	// troubleshooting purposes
	isGlobalAdmin, err := bot.db.IsGlobalAdmin(sender.ID)
	if err != nil {
		bot.logger.WithError(err).Error("can't check if the user is a global admin")
		return
	}

	if isGlobalAdmin {
		var bt = tb.InlineButton{
			Unique: "groups_no_category",
			Text:   "Senza categoria",
		}
		buttons = append(buttons, []tb.InlineButton{bt})

		bot.telebot.Handle(&bt, func(cat botdatabase.ChatCategoryTree) func(callback *tb.Callback) {
			return func(callback *tb.Callback) {
				bot.showCategory(callback.Message, cat, true)
				_ = bot.telebot.Respond(callback)
			}
		}(categoryTree))
	}

	var bt = tb.InlineButton{
		Unique: "groups_list_close",
		Text:   "🚪 Close / Chiudi",
	}
	buttons = append(buttons, []tb.InlineButton{bt})
	bot.telebot.Handle(&bt, func(callback *tb.Callback) {
		_ = bot.telebot.Respond(callback)
		_ = bot.telebot.Delete(callback.Message)
	})

	var sendOptions = tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: buttons,
		},
	}
	msg := "Seleziona il corso di laurea"
	if messageToEdit == nil {
		// No previous messages, send a new one
		_, err = bot.telebot.Send(sender, msg, &sendOptions)
	} else {
		// Previous messages present, edit that one
		_, err = bot.telebot.Edit(messageToEdit, msg, &sendOptions)
	}

	if messageFromUser != nil {
		// We sent the message to the user, however he/she blocked us (or never started a conversation). Send a public
		// message in the group saying that he/she needs to talk in private with the bot first
		if err == tb.ErrNotStartedByUser || err == tb.ErrBlockedByUser {
			replyMessage, _ := bot.telebot.Send(chatToSend, "🇮🇹 Oops, non posso scriverti un messaggio diretto, inizia prima una conversazione diretta con me!\n\n🇬🇧 Oops, I can't text you a direct message, start a direct conversation with me first!",
				&tb.SendOptions{ReplyTo: messageFromUser})

			// Self destruct message in 10s to avoid spamming
			bot.setMessageExpiry(messageFromUser, 10*time.Second)
			bot.setMessageExpiry(replyMessage, 10*time.Second)
		} else if err != nil {
			bot.logger.WithError(err).Warning("can't send group list message to the user")
		} else if !messageFromUser.Private() {
			// The user sent /groups command in a group, however we were able to write him/her in private. Delete the
			// message in the group to avoid spamming
			_ = bot.telebot.Delete(messageFromUser)
		}
	}
}

// showCategory shows the content of the category (e.g. chats associated with this category, and sub categories with
// chats associated to them) by editing the previous message
func (bot *telegramBot) showCategory(m *tb.Message, category botdatabase.ChatCategoryTree, isgeneral bool) {
	msg := strings.Builder{}

	// Show groups in this category before sub-categories
	if len(category.Chats) > 0 {
		for _, v := range category.GetChats() {
			_ = bot.printGroupLinksTelegram(&msg, v)
		}
		msg.WriteString("\n")
	}

	if !isgeneral {
		for _, subcat := range category.GetSubCategoryList() {
			l2cat := category.SubCategories[subcat]

			msg.WriteString("<b>")
			msg.WriteString(subcat)
			msg.WriteString("</b>\n")
			for _, v := range l2cat.GetChats() {
				_ = bot.printGroupLinksTelegram(&msg, v)
			}
			msg.WriteString("\n")
		}
	}

	if msg.Len() == 0 {
		msg.WriteString("Nessun gruppo in questa categoria")
	}

	m, err := bot.telebot.Edit(m, msg.String(), &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
	})
	if err != nil {
		bot.logger.WithError(err).Warning("can't edit message to the user")
	}

	// Delete link list after 10 minutes because invite links can expire, and users have been caught rely on old
	// messages from the bot
	bot.setMessageExpiry(m, 10*time.Minute)
}

// printGroupLinksTelegram formats the group link line in a message (e.g. the line with the group name and the invite
// link). If the group is hidden, this function writes nothing
func (bot *telegramBot) printGroupLinksTelegram(msg *strings.Builder, v *tb.Chat) error {
	settings, err := bot.getChatSettings(v)
	if err != nil {
		bot.logger.WithError(err).WithField("chat", v.ID).Error("Error getting chatroom config")
		return err
	}
	if settings.Hidden {
		return nil
	}

	inviteLink, err := bot.getInviteLink(v)
	if err != nil {
		return err
	}

	msg.WriteString(v.Title)
	msg.WriteString(": ")
	msg.WriteString(fmt.Sprintf("<a href=\"%s\">[ENTRA]</a>", inviteLink))
	msg.WriteString("\n")
	return nil
}

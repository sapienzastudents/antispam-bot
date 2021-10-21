package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v3"
)

// onGroups replies when the user sends /groups or /gruppi command in a private
// char or in a group.
//
// It is just a wrapper to sendGroupListForLinks.
func (bot *telegramBot) onGroups(ctx tb.Context, settings chatSettings) {
	bot.sendGroupListForLinks(ctx.Sender(), nil, ctx.Chat(), ctx.Message())
}

// sendGroupListForLinks sends a list of groups categories as buttons. When the
// user clicks on a button, the message is replaced with the list of groups,
// divided in subcategories.
//
// messageToEdit and messageFromUser can be nil.
func (bot *telegramBot) sendGroupListForLinks(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, messageFromUser *tb.Message) {
	bot.botCommandsRequestsTotal.WithLabelValues("groups").Inc()

	if sender == nil {
		bot.logger.Warn("Have nil on sender, can't send links")
		return
	}
	if chatToSend == nil {
		bot.logger.Warn("Have nil on chatToSend, can't send links")
		return
	}

	categoryTree, err := bot.db.GetChatTree()
	if err != nil {
		bot.logger.WithError(err).Error("Failed to get chatroom list")
		msg, _ := bot.telebot.Send(chatToSend, "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-(")
		bot.setMessageExpiry(msg, 30*time.Second)
		return
	}

	// In Telegram we can add a matrix of InlineButtons. Each row (outer index)
	// can be composed by multiple buttons (inner index). However, categories
	// can be very long, so we stick with one button per row.
	//
	// We support only two layers of "categories": now we draw the first one as
	// a list of buttons. Note that we need to show the right category for the
	// right button, so:
	//
	//	- we register the button for the category using the sha1 of the name
	//	(the whole name might be too long, or contains illegal chars);
	//	- we register a callback handler using a closure to bind the category
	//	variable (so we can show the right subcategory list).
	//
	// Don't try to use the custom "Data" field for buttons here: it doesn't
	// work due some limitations on Telegram side.

	var buttons [][]tb.InlineButton
	for _, category := range categoryTree.GetSubCategoryList() {
		bt := tb.InlineButton{Unique: sha1string(category), Text: category}
		bot.telebot.Handle(&bt, func(cat botdatabase.ChatCategoryTree) tb.HandlerFunc {
			return func(ctx tb.Context) error {
				bot.showCategory(ctx.Callback().Message, cat, false)
				_ = bot.telebot.Respond(ctx.Callback())
				return nil
			}
		}(categoryTree.SubCategories[category]))
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	// Global admins are able to see a special category which contains all
	// groups without a category. This is for troubleshooting purposes.
	isGlobalAdmin, err := bot.db.IsGlobalAdmin(sender.ID)
	if err != nil {
		bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
		return
	}
	if isGlobalAdmin {
		bt := tb.InlineButton{Unique: "groups_no_category", Text: "Senza categoria"}
		bot.telebot.Handle(&bt, func(cat botdatabase.ChatCategoryTree) tb.HandlerFunc {
			return func(ctx tb.Context) error {
				bot.showCategory(ctx.Callback().Message, cat, true)
				_ = bot.telebot.Respond(ctx.Callback())
				return nil
			}
		}(categoryTree))
		buttons = append(buttons, []tb.InlineButton{bt})
	}

	bt := tb.InlineButton{Unique: "groups_list_close", Text: "🚪 Close / Chiudi"}
	bot.telebot.Handle(&bt, func(ctx tb.Context) error {
		_ = bot.telebot.Respond(ctx.Callback())
		_ = bot.telebot.Delete(ctx.Callback().Message)
		return nil
	})
	buttons = append(buttons, []tb.InlineButton{bt})

	// Send (or edit) message with button links.
	sendOptions := &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		DisableWebPagePreview: true,
		ReplyMarkup: &tb.ReplyMarkup{InlineKeyboard: buttons},
	}
	msg := "Seleziona il corso di laurea"
	if messageToEdit == nil {
		// No previous messages, send a new one.
		_, err = bot.telebot.Send(sender, msg, sendOptions)
	} else {
		// Previous messages present, edit that one.
		_, err = bot.telebot.Edit(messageToEdit, msg, sendOptions)
	}
	if messageFromUser != nil {
		if err == tb.ErrNotStartedByUser || err == tb.ErrBlockedByUser {
			// We sent the message to the user, however he blocked us (or never
			// started a conversation). Send a public message in the group
			// saying that he needs to talk in private with the bot first.
			replyMessage, _ := bot.telebot.Send(chatToSend, "🇮🇹 Oops, non posso scriverti un messaggio diretto, inizia prima una conversazione diretta con me!\n\n🇬🇧 Oops, I can't text you a direct message, start a direct conversation with me first!",
				&tb.SendOptions{ReplyTo: messageFromUser})

			// Self destruct messages to avoid spamming.
			bot.setMessageExpiry(messageFromUser, 10*time.Second)
			bot.setMessageExpiry(replyMessage, 10*time.Second)
		} else if err != nil {
			bot.logger.WithError(err).Error("Failed to send group list message to the user")
		} else if !messageFromUser.Private() {
			// The user sent /groups command in a group, however we were able to
			// write him in private. Delete the message in the group to avoid
			// spamming.
			_ = bot.telebot.Delete(messageFromUser)
		}
	}
}

// showCategory shows the content of the given category (e.g. chats associated
// with this category, and subcategories with chats associated to them) by
// editing the previous message.
//
// TODO: Document "isgeneral" parameter.
func (bot *telegramBot) showCategory(m *tb.Message, category botdatabase.ChatCategoryTree, isgeneral bool) {
	msg := strings.Builder{}

	// Show groups in this category before sub-categories.
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
		bot.logger.WithError(err).Warning("Failed to edit message to the user")
	}

	// Force users to ask new invite links because old ones can expire and they
	// can be used accidentally.
	bot.setMessageExpiry(m, 10*time.Minute)
}

// printGroupLinksTelegram formats the group link line in a message (e.g. the
// line with the group name and the invite link) and write the result on msg. If
// the group is hidden, this function writes nothing.
func (bot *telegramBot) printGroupLinksTelegram(msg *strings.Builder, v *tb.Chat) error {
	settings, err := bot.getChatSettings(v)
	if err != nil {
		bot.logger.WithError(err).WithField("chat", v.ID).Error("Failed to get chatroom config")
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

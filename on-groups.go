package main

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func showCategory(m *tb.Message, category botdatabase.ChatCategoryTree, isgeneral bool) {
	msg := strings.Builder{}

	// Show general groups before others
	if len(category.Chats) > 0 {
		for _, v := range category.GetChats() {
			_ = printGroupLinksTelegram(&msg, v)
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
				_ = printGroupLinksTelegram(&msg, v)
			}
			msg.WriteString("\n")
		}
	}

	if msg.Len() == 0 {
		msg.WriteString("Nessun gruppo in questa categoria")
	}

	_, err := b.Edit(m, msg.String(), &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
	})
	if err != nil {
		logger.WithError(err).Warning("can't edit message to the user")
	}
}

func printGroupLinksTelegram(msg *strings.Builder, v *tb.Chat) error {
	settings, err := botdb.GetChatSetting(b, v)
	if err != nil {
		logger.WithError(err).WithField("chat", v.ID).Error("Error getting chatroom config")
		return err
	}
	if settings.Hidden {
		return nil
	}

	chatUUID, err := botdb.GetUUIDFromChat(v.ID)
	if err != nil {
		return err
	}

	msg.WriteString(v.Title)
	msg.WriteString(": ")
	msg.WriteString(fmt.Sprintf("<a href=\"https://telegram.me/%s?start=%s\">[clicca qui e poi premi START]</a>", b.Me.Username, chatUUID.String()))
	msg.WriteString("\n")
	return nil
}

func onGroups(m *tb.Message, _ botdatabase.ChatSettings) {
	sendGroupListForLinks(m.Sender, nil, m.Chat, m)
}

func sendGroupListForLinks(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, messageFromUser *tb.Message) {
	botCommandsRequestsTotal.WithLabelValues("groups").Inc()

	categoryTree, err := botdb.GetChatTree(b)
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
		msg, _ := b.Send(chatToSend, "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-(")
		setMessageExpiration(msg, 30*time.Second)
		return
	}

	var buttons [][]tb.InlineButton

	for _, category := range categoryTree.GetSubCategoryList() {
		var bt = tb.InlineButton{
			Unique: sha1string(category),
			Text:   category,
		}
		buttons = append(buttons, []tb.InlineButton{bt})

		b.Handle(&bt, func(cat botdatabase.ChatCategoryTree) func(callback *tb.Callback) {
			return func(callback *tb.Callback) {
				_ = b.Respond(callback)

				showCategory(callback.Message, cat, false)
			}
		}(categoryTree.SubCategories[category]))
	}

	if botdb.IsGlobalAdmin(sender) {
		var bt = tb.InlineButton{
			Unique: "groups_no_category",
			Text:   "Senza categoria",
		}
		buttons = append(buttons, []tb.InlineButton{bt})

		b.Handle(&bt, func(cat botdatabase.ChatCategoryTree) func(callback *tb.Callback) {
			return func(callback *tb.Callback) {
				_ = b.Respond(callback)

				showCategory(callback.Message, cat, true)
			}
		}(categoryTree))
	}
	var bt = tb.InlineButton{
		Unique: "groups_list_close",
		Text:   "Close / Chiudi",
	}
	buttons = append(buttons, []tb.InlineButton{bt})
	b.Handle(&bt, func(callback *tb.Callback) {
		_ = b.Respond(callback)
		_ = b.Delete(callback.Message)
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
		_, err = b.Send(sender, msg, &sendOptions)
	} else {
		_, err = b.Edit(messageToEdit, msg, &sendOptions)
	}
	if messageFromUser != nil {
		if err == tb.ErrNotStartedByUser || err == tb.ErrBlockedByUser {
			replyMessage, _ := b.Send(chatToSend, "🇮🇹 Oops, non posso scriverti un messaggio diretto, inizia prima una conversazione diretta con me!\n\n🇬🇧 Oops, I can't text you a direct message, start a direct conversation with me first!",
				&tb.SendOptions{ReplyTo: messageFromUser})

			// Self destruct message in 10s
			setMessageExpiration(messageFromUser, 10*time.Second)
			setMessageExpiration(replyMessage, 10*time.Second)
		} else if err != nil {
			logger.WithError(err).Warning("can't send group list message to the user")
		} else if !messageFromUser.Private() {
			// User contacted in private before, and command in a public group -> remove user public messages
			_ = b.Delete(messageFromUser)
		}
	}
}

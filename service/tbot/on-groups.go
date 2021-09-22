package tbot

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) showCategory(m *tb.Message, category botdatabase.ChatCategoryTree, isgeneral bool) {
	msg := strings.Builder{}

	// Show general groups before others
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

	// Delete link list after 10 minutes because invite links will expire soon
	bot.setMessageExpiry(m, 10*time.Minute)
}

func (bot *telegramBot) printGroupLinksTelegram(msg *strings.Builder, v *tb.Chat) error {
	settings, err := bot.db.GetChatSetting(bot.telebot, v)
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

func (bot *telegramBot) onGroups(m *tb.Message, _ chatSettings) {
	bot.sendGroupListForLinks(m.Sender, nil, m.Chat, m)
}

func (bot *telegramBot) sendGroupListForLinks(sender *tb.User, messageToEdit *tb.Message, chatToSend *tb.Chat, messageFromUser *tb.Message) {
	bot.botCommandsRequestsTotal.WithLabelValues("groups").Inc()

	categoryTree, err := bot.db.GetChatTree(bot.telebot)
	if err != nil {
		bot.logger.WithError(err).Error("Error getting chatroom list")
		msg, _ := bot.telebot.Send(chatToSend, "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-(")
		bot.setMessageExpiry(msg, 30*time.Second)
		return
	}

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

	if bot.db.IsGlobalAdmin(sender.ID) {
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
		Text:   "Close / Chiudi",
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
		_, err = bot.telebot.Send(sender, msg, &sendOptions)
	} else {
		_, err = bot.telebot.Edit(messageToEdit, msg, &sendOptions)
	}
	if messageFromUser != nil {
		if err == tb.ErrNotStartedByUser || err == tb.ErrBlockedByUser {
			replyMessage, _ := bot.telebot.Send(chatToSend, "🇮🇹 Oops, non posso scriverti un messaggio diretto, inizia prima una conversazione diretta con me!\n\n🇬🇧 Oops, I can't text you a direct message, start a direct conversation with me first!",
				&tb.SendOptions{ReplyTo: messageFromUser})

			// Self destruct message in 10s
			bot.setMessageExpiry(messageFromUser, 10*time.Second)
			bot.setMessageExpiry(replyMessage, 10*time.Second)
		} else if err != nil {
			bot.logger.WithError(err).Warning("can't send group list message to the user")
		} else if !messageFromUser.Private() {
			// User contacted in private before, and command in a public group -> remove user public messages
			_ = bot.telebot.Delete(messageFromUser)
		}
	}
}

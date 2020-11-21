package main

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
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

	_, err := b.Edit(m, msg.String(), &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
	})
	if err != nil {
		logger.Warning("can't send message to the user ", err)
	}
}

func printGroupLinksTelegram(msg *strings.Builder, v *tb.Chat) error {
	settings, err := botdb.GetChatSetting(b, v)
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom config")
		return err
	}
	if settings.Hidden {
		return nil
	}

	if v.InviteLink == "" {
		v.InviteLink, err = b.GetInviteLink(v)

		if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
			apierr, _ := err.(*tb.APIError)
			v, err = b.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
			if err != nil {
				logger.Warning("can't get chat info ", err)
				return err
			}

			v.InviteLink, err = b.GetInviteLink(v)
			if err != nil {
				logger.Warning("can't get invite link ", err)
				return err
			}
		} else if err != nil {
			logger.Warning("can't get chat info ", err)
			return err
		}
		_ = botdb.UpdateMyChatroomList(v)
	}

	msg.WriteString(v.Title)
	msg.WriteString(": ")
	msg.WriteString(v.InviteLink)
	msg.WriteString("\n")
	return nil
}

func onGroups(m *tb.Message, _ botdatabase.ChatSettings) {
	categoryTree, err := botdb.GetChatTree(b)
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
		_, _ = b.Send(m.Chat, "Ooops, ho perso qualche rotella, avverti il mio admin che mi sono rotto :-(")
		return
	}

	var buttons [][]tb.InlineButton

	for _, category := range categoryTree.GetSubCategoryList() {
		var bt = tb.InlineButton{
			Unique: Sha1(category),
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

	if len(categoryTree.Chats) > 0 {
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

	_, err = b.Send(m.Sender, "Seleziona il corso di laurea", &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		DisableWebPagePreview: true,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: buttons,
		},
	})
	if err != nil {
		logger.Warning("can't send message to the user ", err)
	}

	if !m.Private() {
		_, _ = b.Send(m.Chat, "🇮🇹 Ti ho scritto in privato!\n\n🇬🇧 I sent you a direct message!", &tb.SendOptions{ReplyTo: m})
	}
}

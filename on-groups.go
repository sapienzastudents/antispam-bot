package main

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"sort"
	"strings"
)

const ChatCategorySeparator = "||"

func showCategory(m *tb.Message, category string) {
	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
	} else {

		msg := strings.Builder{}
		subCategoriesChat := map[string][]*tb.Chat{}
		subCategories := Set{}
		for _, v := range chatrooms {
			settings, err := botdb.GetChatSetting(b, v)
			if err != nil {
				logger.WithError(err).Error("Error getting chatroom config")
				continue
			}
			if settings.Hidden {
				continue
			}
			chatCategory, err := botdb.GetChatCategory(v)
			if err != nil {
				logger.WithError(err).WithField("chat", v.ID).Error("Error getting chatroom category")
				continue
			}
			chatCategoryLevels := strings.Split(chatCategory, ChatCategorySeparator)
			if category != chatCategoryLevels[0] {
				continue
			}
			subCategory := ""
			if len(chatCategoryLevels) > 1 {
				subCategory = chatCategoryLevels[1]
			}

			if v.InviteLink == "" {
				v.InviteLink, err = b.GetInviteLink(v)

				if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
					apierr, _ := err.(*tb.APIError)
					v, err = b.ChatByID(fmt.Sprint(apierr.Parameters["migrate_to_chat_id"]))
					if err != nil {
						logger.Warning("can't get chat info ", err)
						continue
					}

					v.InviteLink, err = b.GetInviteLink(v)
					if err != nil {
						logger.Warning("can't get invite link ", err)
						continue
					}
				} else if err != nil {
					logger.Warning("can't get chat info ", err)
					continue
				}
				_ = botdb.UpdateMyChatroomList(v)
			}

			subCategoriesChat[subCategory] = append(subCategoriesChat[subCategory], v)
			subCategories.Add(subCategory)
		}

		// Show general groups before others
		if groups, ok := subCategoriesChat[""]; ok {
			sort.Slice(groups, func(i, j int) bool {
				return groups[i].Title < groups[j].Title
			})
			for _, v := range groups {
				msg.WriteString(v.Title)
				msg.WriteString(": ")
				msg.WriteString(v.InviteLink)
				msg.WriteString("\n")
			}
			msg.WriteString("\n")
		}

		for _, subcat := range subCategories.GetAsOrderedList() {
			if subcat == "" {
				continue
			}
			groups := subCategoriesChat[subcat]

			sort.Slice(groups, func(i, j int) bool {
				return groups[i].Title < groups[j].Title
			})
			msg.WriteString("<b>")
			msg.WriteString(subcat)
			msg.WriteString("</b>\n")
			for _, v := range groups {
				msg.WriteString(v.Title)
				msg.WriteString(": ")
				msg.WriteString(v.InviteLink)
				msg.WriteString("\n")
			}
			msg.WriteString("\n")
		}

		_, err = b.Edit(m, msg.String(), &tb.SendOptions{
			ParseMode:             tb.ModeHTML,
			DisableWebPagePreview: true,
		})
		if err != nil {
			logger.Warning("can't send message to the user ", err)
		}
	}
}

func onGroups(m *tb.Message, _ botdatabase.ChatSettings) {
	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
	} else {
		categories := Set{}
		for _, v := range chatrooms {
			settings, err := botdb.GetChatSetting(b, v)
			if err != nil {
				logger.WithError(err).Error("Error getting chatroom config")
				continue
			}
			if settings.Hidden {
				continue
			}
			category, err := botdb.GetChatCategory(v)
			if err != nil {
				logger.WithError(err).Error("Can't get chat category")
				continue
			}
			categoryLevels := strings.Split(category, ChatCategorySeparator)
			categories.Add(categoryLevels[0])
		}

		var buttons [][]tb.InlineButton
		for _, category := range categories.GetAsOrderedList() {
			label := category
			if label == "" {
				label = "Gruppi generali Sapienza"
			}
			var bt = tb.InlineButton{
				Unique: Sha1(category),
				Text:   label,
			}
			buttons = append(buttons, []tb.InlineButton{bt})

			b.Handle(&bt, func(cat string) func(callback *tb.Callback) {
				return func(callback *tb.Callback) {
					_ = b.Respond(callback)

					showCategory(callback.Message, cat)
				}
			}(category))
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
	}

	if !m.Private() {
		_, _ = b.Send(m.Chat, "🇮🇹 Ti ho scritto in privato!\n\n🇬🇧 I sent you a direct message!", &tb.SendOptions{ReplyTo: m})
	}
}

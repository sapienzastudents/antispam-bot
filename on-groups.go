package main

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"sort"
	"strings"
)

func onGroups(m *tb.Message, _ botdatabase.ChatSettings) {
	logger.Debugf("My chat room requested by %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName)

	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
	} else {
		sort.Slice(chatrooms, func(i, j int) bool {
			return chatrooms[i].Title < chatrooms[j].Title
		})

		msg := strings.Builder{}
		for _, v := range chatrooms {
			settings, err := botdb.GetChatSetting(b, v)
			if err != nil {
				logger.WithError(err).Error("Error getting chatroom config")
				continue
			}
			if settings.Hidden {
				continue
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

			msg.WriteString("<b>")
			msg.WriteString(v.Title)
			msg.WriteString("</b>\n")
			msg.WriteString(v.InviteLink)
			msg.WriteString("\n\n")
		}

		_, err = b.Send(m.Sender, msg.String(), &tb.SendOptions{
			ParseMode:             tb.ModeHTML,
			DisableWebPagePreview: true,
		})
		if err != nil {
			logger.Warning("can't send message to the user ", err)
		}
	}

	if !m.Private() {
		_, _ = b.Send(m.Chat, "🇮🇹 Ti ho scritto in privato!\n\n🇬🇧 I sent you a direct message!", &tb.SendOptions{ReplyTo: m})
	}
}

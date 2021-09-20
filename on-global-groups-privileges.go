package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onGroupsPrivileges(m *tb.Message, _ botdatabase.ChatSettings) {
	logger.Debugf("My chat room privileges requested by %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName)

	waitingmsg, _ := b.Send(m.Chat, "Work in progress...")

	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
	} else {
		sort.Slice(chatrooms, func(i, j int) bool {
			return chatrooms[i].Title < chatrooms[j].Title
		})

		msg := strings.Builder{}
		for k, v := range botPermissionsTag {
			msg.WriteString(k)
			msg.WriteString(" -> ")
			msg.WriteString(v)
			msg.WriteString("\n")
		}
		msg.WriteString("\n")

		for _, v := range chatrooms {
			newInfos, err := b.ChatByID(fmt.Sprint(v.ID))
			if err != nil {
				logger.Warning("can't get refreshed infos for chatroom ", v, " ", err)
				continue
			}
			v = newInfos

			me, err := b.ChatMemberOf(v, b.Me)
			if err != nil {
				logger.Warning("can't get refreshed infos for chatroom ", v, " ", err)
				continue
			}

			msg.WriteString(" - ")
			msg.WriteString(v.Title)
			msg.WriteString(" : ")
			if me.Role != tb.Administrator {
				msg.WriteString("❌ not admin\n")
			} else {
				var missingPrivileges = synthetizePrivileges(me)
				if len(missingPrivileges) == 0 {
					msg.WriteString("✅\n")
				} else {
					for _, k := range missingPrivileges {
						msg.WriteString(botPermissionsTag[k])
					}
					msg.WriteString("❌\n")
				}
			}

			_, err = b.Edit(waitingmsg, msg.String())
			if err != nil {
				logger.Warning("[global] can't edit message to the user ", err)
			}

			// Do not trigger Telegram rate limit
			time.Sleep(500 * time.Millisecond)
		}

		msg.WriteString("\ndone")

		_, err = b.Edit(waitingmsg, msg.String())
		if err != nil {
			logger.Warning("[global] can't edit final message to the user ", err)
		}
	}
}

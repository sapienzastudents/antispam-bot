package main

import (
	"fmt"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"reflect"
	"sort"
	"strings"
	"time"
)

func onGroupsPrivileges(m *tb.Message, _ botdatabase.ChatSettings) {
	logger.Debugf("My chat room privileges requested by %d (%s %s %s)", m.Sender.ID, m.Sender.Username, m.Sender.FirstName, m.Sender.LastName)

	var botpermissions = map[string]string{
		"can_change_info":      "C",
		"can_delete_messages":  "D",
		"can_invite_users":     "I",
		"can_restrict_members": "R",
		"can_pin_messages":     "N",
		"can_promote_members":  "P",
	}

	waitingmsg, _ := b.Send(m.Chat, "Work in progress...")

	chatrooms, err := botdb.ListMyChatrooms()
	if err != nil {
		logger.WithError(err).Error("Error getting chatroom list")
	} else {
		sort.Slice(chatrooms, func(i, j int) bool {
			return chatrooms[i].Title < chatrooms[j].Title
		})

		msg := strings.Builder{}
		for k, v := range botpermissions {
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
				var permissionsOk = true

				t := reflect.TypeOf(me.Rights)
				right := reflect.ValueOf(me.Rights)
				for i := 0; i < t.NumField(); i++ {
					k := t.Field(i).Tag.Get("json")
					tag, ok := botpermissions[k]
					if !ok {
						// Skip this field
						continue
					}

					f := right.Field(i)
					if !f.Bool() {
						msg.WriteString(tag)
						permissionsOk = false
					}
				}

				if permissionsOk {
					msg.WriteString("✅\n")
				} else {
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

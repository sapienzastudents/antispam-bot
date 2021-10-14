package tbot

import (
	"sort"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v3"
)

// onGroupsPrivileges sends a list of all chats and relative permission on
// /groupscheck command. Used for dbug purposes only.
func (bot *telegramBot) onGroupsPrivileges(ctx tb.Context, settings chatSettings) {
	m := ctx.Message()
	if m == nil {
		bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
		return
	}

	bot.logger.WithFields(logrus.Fields{
		"userid": m.Sender.ID,
		"userusername": m.Sender.Username,
		"userfirstname": m.Sender.FirstName,
		"userlastname": m.Sender.LastName,
	}).Debug("Chat room list with privileges requested by user")

	// The list with all chats and relative permissions takes time to build, so
	// send an "ack" message, it will be edited at the end.
	waitingmsg, err := bot.telebot.Send(m.Chat, "Work in progress...")
	if err != nil {
		bot.logger.WithError(err).Error("Failed to reply on /groupscheck")
	}

	chatrooms, err := bot.db.ListMyChatrooms()
	if err != nil {
		bot.logger.WithError(err).Error("Failed to get chatroom list")
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
			newInfos, err := bot.telebot.ChatByID(v.ID)
			if err != nil {
				bot.logger.WithError(err).WithField("chat", v).Warn("Failed to get refreshed infos for chatroom")
				continue
			}
			v = newInfos

			me, err := bot.telebot.ChatMemberOf(v, bot.telebot.Me)
			if err != nil {
				bot.logger.WithError(err).WithField("chat", v).Warn("Failed to get refreshed infos for chatroom")
				continue
			}

			msg.WriteString(" - ")
			msg.WriteString(v.Title)
			msg.WriteString(" : ")
			if me.Role != tb.Administrator {
				msg.WriteString("❌ not admin\n")
			} else {
				missingPrivileges := synthetizePrivileges(me)
				if len(missingPrivileges) == 0 {
					msg.WriteString("✅\n")
				} else {
					for _, k := range missingPrivileges {
						msg.WriteString(botPermissionsTag[k])
					}
					msg.WriteString("❌\n")
				}
			}

			_, err = bot.telebot.Edit(waitingmsg, msg.String())
			if err != nil {
				bot.logger.Warning("[global] can't edit message to the user ", err)
			}

			// Do not trigger Telegram rate limit
			time.Sleep(500 * time.Millisecond)
		}

		msg.WriteString("\ndone")

		_, err = bot.telebot.Edit(waitingmsg, msg.String())
		if err != nil {
			bot.logger.Warning("[global] can't edit final message to the user ", err)
		}
	}
}

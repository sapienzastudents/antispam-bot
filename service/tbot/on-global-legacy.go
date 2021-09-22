package tbot

import (
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (bot *telegramBot) onCut(m *tb.Message, _ chatSettings) {
	if !m.Private() {
		return
	}

	parts := strings.SplitN(m.Payload, " ", 2)
	if len(parts) != 2 {
		_, _ = bot.telebot.Send(m.Chat, "missing chat ID and/or message ID")
		return
	}

	chatId, err := bot.telebot.ChatByID(parts[1])
	if err != nil {
		_, _ = bot.telebot.Send(m.Chat, "Invalid chat ID/name specified")
		return
	}

	messageId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		_, _ = bot.telebot.Send(m.Chat, "Invalid message ID specified")
		return
	}

	err = bot.telebot.Delete(&tb.Message{Chat: chatId, ID: int(messageId)})
	if err != nil {
		_, _ = bot.telebot.Send(m.Chat, "Error deleting message: "+err.Error())
	} else {
		_, _ = bot.telebot.Send(m.Chat, "Message deleted")
	}
}

func (bot *telegramBot) onEmergencyRemove(m *tb.Message, _ chatSettings) {
	err := bot.telebot.Delete(m)
	if err != nil {
		bot.logger.Error("Can't delete messages ", err)
		return
	}

	if m.ReplyTo != nil {
		err = bot.telebot.Delete(m.ReplyTo)
		if err != nil {
			bot.logger.Error("Can't delete messages ", err)
			return
		}
	}
}

func (bot *telegramBot) onEmergencyElevate(m *tb.Message, _ chatSettings) {
	err := bot.telebot.Delete(m)
	if err != nil {
		bot.logger.Error("Can't delete messages ", err)
		return
	}

	member, err := bot.telebot.ChatMemberOf(m.Chat, m.Sender)
	if err != nil {
		bot.logger.Error("Can't get member of ", err)
	} else {
		member.CanDeleteMessages = true
		member.CanChangeInfo = true
		member.CanInviteUsers = true
		member.CanPinMessages = true
		member.CanRestrictMembers = true
		member.CanPromoteMembers = true
		err = bot.telebot.Promote(m.Chat, member)
		if err != nil {
			bot.logger.Error("Can't elevate ", err)
		}
	}
}

func (bot *telegramBot) onEmergencyReduce(m *tb.Message, _ chatSettings) {
	err := bot.telebot.Delete(m)
	if err != nil {
		bot.logger.Error("Can't delete messages ", err)
		return
	}

	member, err := bot.telebot.ChatMemberOf(m.Chat, m.Sender)
	if err != nil {
		bot.logger.Error("Can't get member of ", err)
	} else {
		member.CanDeleteMessages = false
		member.CanChangeInfo = false
		member.CanInviteUsers = false
		member.CanPinMessages = false
		member.CanRestrictMembers = false
		member.CanPromoteMembers = false
		err = bot.telebot.Promote(m.Chat, member)
		if err != nil {
			bot.logger.Error("Can't reduce ", err)
		}
	}
}

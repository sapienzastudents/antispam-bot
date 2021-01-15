package main

import (
	"fmt"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onEmergencyRemove(m *tb.Message, _ botdatabase.ChatSettings) {
	err := b.Delete(m.ReplyTo)
	if err != nil {
		logger.Error("Can't delete messages ", err)
		return
	}
	err = b.Delete(m)
	if err != nil {
		logger.Error("Can't delete messages ", err)
		return
	}
}

func onEmergencyElevate(m *tb.Message, _ botdatabase.ChatSettings) {
	err := b.Delete(m)
	if err != nil {
		logger.Error("Can't delete messages ", err)
		return
	}

	member, err := b.ChatMemberOf(m.Chat, m.Sender)
	if err != nil {
		logger.Error("Can't get member of ", err)
	} else {
		member.CanDeleteMessages = true
		member.CanChangeInfo = true
		member.CanInviteUsers = true
		member.CanPinMessages = true
		member.CanRestrictMembers = true
		member.CanPromoteMembers = true
		err = b.Promote(m.Chat, member)
		if err != nil {
			logger.Error("Can't elevate ", err)
		}
	}
}

func onEmergencyReduce(m *tb.Message, _ botdatabase.ChatSettings) {
	err := b.Delete(m)
	if err != nil {
		logger.Error("Can't delete messages ", err)
		return
	}

	member, err := b.ChatMemberOf(m.Chat, m.Sender)
	if err != nil {
		logger.Error("Can't get member of ", err)
	} else {
		member.CanDeleteMessages = false
		member.CanChangeInfo = false
		member.CanInviteUsers = false
		member.CanPinMessages = false
		member.CanRestrictMembers = false
		member.CanPromoteMembers = false
		err = b.Promote(m.Chat, member)
		if err != nil {
			logger.Error("Can't reduce ", err)
		}
	}
}

func onSigHup(m *tb.Message, _ botdatabase.ChatSettings) {
	err := botdb.DoCacheUpdate(b, groupUserCount)
	if err != nil {
		logger.WithError(err).Warning("can't handle sighup / refresh data")
		_, _ = b.Send(m.Chat, "Errore: "+err.Error())
	} else {
		_, _ = b.Send(m.Chat, "Reload OK")
	}
}

func onVersion(m *tb.Message, _ botdatabase.ChatSettings) {
	msg := fmt.Sprintf("Version %s", AppVersion)
	_, _ = b.Send(m.Chat, msg)
}

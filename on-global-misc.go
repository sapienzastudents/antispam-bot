package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onEmergencyRemove(m *tb.Message, _ ChatSettings) {
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

func onEmergencyElevate(m *tb.Message, _ ChatSettings) {
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

func onSigHup(m *tb.Message, _ ChatSettings) {
	err := botdb.DoCacheUpdate()
	if err != nil {
		b.Send(m.Chat, "Errore: "+err.Error())
	} else {
		b.Send(m.Chat, "Reload OK")
	}
}

func onVersion(m *tb.Message, _ ChatSettings) {
	msg := fmt.Sprintf("Version %s", APP_VERSION)
	_, _ = b.Send(m.Chat, msg)
}

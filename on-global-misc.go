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

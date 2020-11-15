package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onSigDel(m *tb.Message, _ ChatSettings) {
	if !botdb.IsGlobalAdmin(m.Sender) {
		return
	}

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

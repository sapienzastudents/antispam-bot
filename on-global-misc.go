package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

func onGLine(m *tb.Message, _ botdatabase.ChatSettings) {
	_ = b.Delete(m)
	if m.Sender.IsBot || (m.ReplyTo != nil && m.ReplyTo.Sender != nil && m.ReplyTo.Sender.IsBot) {
		return
	} else if m.ReplyTo != nil && m.ReplyTo.Sender != nil {
		if botdb.IsGlobalAdmin(m.ReplyTo.Sender) {
			logger.WithField("chatid", m.Chat.ID).Warn("Won't g-line a global admin")
			return
		}
		_ = b.Delete(m.ReplyTo)
		banUser(m.Chat, m.ReplyTo.Sender)
		botdb.SetUserBanned(int64(m.ReplyTo.Sender.ID))
		_, _ = b.Send(m.Sender, fmt.Sprint("GLine ok for ", m.ReplyTo.Sender))
		logger.WithFields(logrus.Fields{
			"chatid":     m.Chat.ID,
			"adminid":    m.Sender.ID,
			"targetuser": m.ReplyTo.Sender.ID,
		}).Info("g-line user")
	} else if m.Text != "" {
		payload := strings.TrimSpace(m.Text)
		if strings.ContainsRune(payload, ' ') {
			parts := strings.Split(m.Text, " ")
			userID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				_, _ = b.Send(m.Chat, "Invalid ID specified")
				return
			}
			if botdb.IsGlobalAdmin(&tb.User{ID: int(userID)}) {
				logger.WithField("chatid", m.Chat.ID).Warn("Won't g-line a global admin")
				return
			}
			botdb.SetUserBanned(userID)
			_, _ = b.Send(m.Sender, fmt.Sprint("GLine ok for ", userID))
			logger.WithFields(logrus.Fields{
				"chatid":     m.Chat.ID,
				"adminid":    m.Sender.ID,
				"targetuser": userID,
			}).Info("g-line user")
		}
	}
}

func onEmergencyRemove(m *tb.Message, _ botdatabase.ChatSettings) {
	err := b.Delete(m)
	if err != nil {
		logger.Error("Can't delete messages ", err)
		return
	}

	if m.ReplyTo != nil {
		err = b.Delete(m.ReplyTo)
		if err != nil {
			logger.Error("Can't delete messages ", err)
			return
		}
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

func onSigTerm(m *tb.Message, _ botdatabase.ChatSettings) {
	if !m.Private() {
		_ = b.Delete(m)
		err := botdb.DeleteChat(m.Chat.ID)
		if err != nil {
			logger.WithError(err).Error("can't delete chat info from redis")
			return
		}
		err = b.Leave(m.Chat)
		if err != nil {
			logger.WithError(err).Error("can't leave chat")
			return
		}
	}
}

func onVersion(m *tb.Message, _ botdatabase.ChatSettings) {
	msg := fmt.Sprintf("Version %s", AppVersion)
	_, _ = b.Send(m.Chat, msg)
}

package main

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func onUserJoined(m *tb.Message, settings ChatSettings) {
	if m.IsService() && !m.Private() && m.UserJoined.ID == b.Me.ID {
		logger.Infof("Joining chat %s", m.Chat.Title)
		return
	}

	logger.Debugf("User %d (%s %s %s) joined chat %s (%d)", m.UserJoined.ID, m.UserJoined.Username,
		m.UserJoined.FirstName, m.UserJoined.LastName, m.Chat.Title, m.Chat.ID)

	if botdb.IsGlobalAdmin(m.UserJoined) {
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

	if settings.OnBlacklistCAS.Action != ACTION_NONE && settings.OnBlacklistCAS.Action != ACTION_DELETE_MSG && IsCASBanned(m.Sender.ID) {
		logger.Infof("User %d CAS-banned, performing action: %s", m.Sender.ID, prettyActionName(settings.OnBlacklistCAS))
		performAction(m, m.Sender, settings.OnBlacklistCAS)
		return
	}

	// Note: nothing personal. We were forced to write these blocks for chinese texts in a period of time when bots were
	// targetting our group. This check is trying to avoid banning people randomly just for having chinese/arabic names,
	// however false positive might arise
	textvalues := []string{
		m.UserJoined.Username,
		m.UserJoined.FirstName,
		m.UserJoined.LastName,
	}

	for _, text := range textvalues {
		if settings.OnJoinChinese.Action != ACTION_NONE {
			chinesePercent := chineseChars(text)
			logger.Debugf("SPAM detection (%s): chinese %f", text, chinesePercent)
			if chinesePercent > 0.5 {
				performAction(m, m.UserJoined, settings.OnJoinChinese)
				return
			}
		}

		if settings.OnJoinArabic.Action != ACTION_NONE {
			arabicPercent := arabicChars(text)
			logger.Debugf("SPAM detection (%s): arabic %f", text, arabicPercent)
			if arabicPercent > 0.5 {
				performAction(m, m.UserJoined, settings.OnJoinArabic)
				return
			}
		}
	}

	if settings.OnJoinDelete {
		err := b.Delete(m)
		if err != nil {
			logger.Critical("Cannot delete join message: ", err)
		}
	}
}

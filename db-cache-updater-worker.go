package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

func (db *_botDatabase) DoCacheUpdate() error {
	startms := time.Now()
	logger.Info("Chat admin scan start")

	chats, err := db.ListMyChatrooms()
	if err != nil {
		return err
	}

	for _, chat := range chats {
		logger.Infof("Scanning chat %d %s", chat.ID, chat.Title)

		newChatInfo, err := b.ChatByID(fmt.Sprint(chat.ID))
		if err != nil {
			if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == 400 || apierr.Code == 403) {
				logger.Criticalf("Chat %s not found, removing configuration", chat.Title)
				db.LeftChatroom(chat)
				continue
			}
			logger.Criticalf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error())
			continue
		}
		chat = newChatInfo
		logger.Infof("New chat info: %d %s", chat.ID, chat.Title)

		_, err = b.GetInviteLink(chat)
		if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
			// We have both the old group and the new group, remove the old one only
			botdb.LeftChatroom(chat)
			continue
		}

		admins, err := b.AdminsOf(chat)
		if err != nil {
			if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == 400 || apierr.Code == 403) {
				logger.Criticalf("Chat %s not found, removing configuration", chat.Title)
				db.LeftChatroom(chat)
				continue
			}
			logger.Criticalf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error())
			continue
		}

		chatsettings, err := db.GetChatSetting(chat)
		if err != nil {
			logger.Criticalf("Cannot get chat settings for chat %d %s: %s", chat.ID, chat.Title, err.Error())
			continue
		}

		chatsettings.ChatAdmins = admins
		err = db.SetChatSettings(chat, chatsettings)
		if err != nil {
			logger.Criticalf("Cannot save chat settings for chat %d %s: %s", chat.ID, chat.Title, err.Error())
		}
	}

	logger.Infof("Chat admin scan done in %.3f seconds", time.Now().Sub(startms).Seconds())
	return nil
}

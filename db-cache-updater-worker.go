package main

import (
	"fmt"
	"github.com/pkg/errors"
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
		err = db.DoCacheUpdateForChat(chat)
		if err != nil {
			logger.Warning("Error updating chat ", chat.Title, " ", err.Error())
		}

		// Do not ask too quickly
		time.Sleep(1 * time.Second)
	}

	logger.Infof("Chat admin scan done in %.3f seconds", time.Now().Sub(startms).Seconds())
	return nil
}

func (db *_botDatabase) DoCacheUpdateForChat(chat *tb.Chat) error {
	newChatInfo, err := b.ChatByID(fmt.Sprint(chat.ID))
	if err != nil {
		if apierr, ok := err.(*tb.APIError); ok && (apierr.Code == 400 || apierr.Code == 403) {
			_ = db.LeftChatroom(chat)
			return errors.Wrap(err, fmt.Sprintf("Chat %s not found, removing configuration", chat.Title))
		}
		return errors.Wrap(err, fmt.Sprintf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error()))
	}
	chat = newChatInfo

	inviteLink, err := b.GetInviteLink(chat)
	if err != nil && err.Error() == tb.ErrGroupMigrated.Error() {
		// We have both the old group and the new group, remove the old one only
		return botdb.LeftChatroom(chat)
	}
	chat.InviteLink = inviteLink

	admins, err := b.AdminsOf(chat)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error()))
	}

	chatsettings, err := db.GetChatSetting(chat)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Cannot get chat settings for chat %d %s: %s", chat.ID, chat.Title, err.Error()))
	}

	chatsettings.ChatAdmins.SetFromChat(admins)
	err = db.SetChatSettings(chat, chatsettings)
	if err != nil {
		logger.Criticalf("Cannot save chat settings for chat %d %s: %s", chat.ID, chat.Title, err.Error())
	}

	return db.UpdateMyChatroomList(chat)
}

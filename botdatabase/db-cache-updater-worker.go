package botdatabase

import (
	"fmt"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

func (db *_botDatabase) DoCacheUpdate(b *tb.Bot) error {
	startms := time.Now()
	db.logger.Info("Chat admin scan start")

	chats, err := db.ListMyChatrooms()
	if err != nil {
		return err
	}

	for _, chat := range chats {
		err = db.DoCacheUpdateForChat(b, chat)
		if err != nil {
			db.logger.WithError(err).WithField("chat_id", chat.ID).Warning("Error updating chat ", chat.Title)
		}

		// Do not ask too quickly
		time.Sleep(1 * time.Second)
	}

	db.logger.Infof("Chat admin scan done in %.3f seconds", time.Since(startms).Seconds())
	return nil
}

func (db *_botDatabase) DoCacheUpdateForChat(b *tb.Bot, chat *tb.Chat) error {
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
		return db.LeftChatroom(chat)
	}
	chat.InviteLink = inviteLink

	admins, err := b.AdminsOf(chat)
	if err != nil {
		db.logger.WithError(err).WithField("chat_id", chat.ID).Error("Error getting admins for chat ", chat.Title)
		return errors.Wrap(err, fmt.Sprintf("Error getting admins for chat %d (%s): %s", chat.ID, chat.Title, err.Error()))
	}

	chatsettings, err := db.GetChatSetting(b, chat)
	if err != nil {
		db.logger.WithError(err).WithField("chat_id", chat.ID).Error("Cannot get chat settings for chat ", chat.Title)
		return errors.Wrap(err, fmt.Sprintf("Cannot get chat settings for chat %d %s: %s", chat.ID, chat.Title, err.Error()))
	}

	chatsettings.ChatAdmins.SetFromChat(admins)
	err = db.SetChatSettings(chat, chatsettings)
	if err != nil {
		db.logger.WithError(err).WithField("chat_id", chat.ID).Error("Cannot save chat settings for chat ", chat.Title)
		return err
	}

	return db.UpdateMyChatroomList(chat)
}
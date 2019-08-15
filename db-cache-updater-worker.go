package main

import "time"

func (db *_botDatabase) DoCacheUpdate() error {
	startms := time.Now()
	logger.Info("Chat admin scan start")

	chats, err := db.ListMyChatrooms()
	if err != nil {
		return err
	}

	for _, chat := range chats {
		admins, err := b.AdminsOf(chat)
		if err != nil {
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

package botdatabase

import "fmt"

func (db *_botDatabase) DeleteChat(chatID int64) error {
	err := db.redisconn.HDel("public-links", fmt.Sprint(chatID)).Err()
	if err != nil {
		return err
	}
	err = db.redisconn.HDel("settings", fmt.Sprintf("%d", chatID)).Err()
	if err != nil {
		return err
	}
	return db.redisconn.HDel("chatrooms", fmt.Sprintf("%d", chatID)).Err()
}

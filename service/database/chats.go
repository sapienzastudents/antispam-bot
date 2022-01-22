package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v3"
)

// migrateOldChats migrates old tracked chats on the database to the new
// structure.
//
// TODO: This method can be removed in the future.
func (db *Database) migrateOldChats() error {
	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.conn.HScan(context.TODO(), "chatrooms", cursor, "", -1).Result()
		if errors.Is(err, redis.Nil) {
			return nil // Old key doesn't exist, nothing to migrate.
		} else if err != nil {
			return fmt.Errorf("on scanning \"chatrooms\": %w", err)
		}

		for i := 0; i < len(keys); i += 2 {
			type ChatTmp struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
			}

			chat := ChatTmp{}
			err = json.Unmarshal([]byte(keys[i+1]), &chat)
			if err != nil {
				return fmt.Errorf("on unmarshalling old chatroom %q: %w", keys[i+1], err)
			}

			if err := db.AddChat(&tb.Chat{ID: chat.ID, Title: chat.Title}); err != nil {
				return fmt.Errorf("on adding %d chat: %w", chat.ID, err)
			}
		}

		// SCAN cycle end
		if cursor == 0 {
			break
		}
	}

	// Delete old key, so the next call will skip the migration.
	if err := db.conn.Del(context.TODO(), "chatrooms").Err(); err != nil {
		return fmt.Errorf("on deleting \"chatrooms\": %w", err)
	}
	return nil
}

// AddChat adds or updated the given chat into the DB.
//
// As Telegram doesn't offer a way to track in which chatrooms the bot is, we
// need to store it in Redis.
//
// Only ID and Title fields in tb.Chat are saved into the DB.
func (db *Database) AddChat(c *tb.Chat) error {
	// First add the given chat ID as tracked chats.
	id := strconv.FormatInt(c.ID, 10)
	if err := db.conn.SAdd(context.TODO(), "chats", id).Err(); err != nil {
		return fmt.Errorf("on adding/updating the given chat to \"chats\" hash set: %w", err)
	}

	// Then save chat's details.
	hid := "chats:" + id
	if err := db.conn.HSet(context.TODO(), hid, "title", c.Title).Err(); err != nil {
		return fmt.Errorf("on adding/updating the given chat title to %q hash: %w", hid, err)
	}

	return nil
}

// DeleteChat removes the chat info of the given chat ID.
//
// If the given chat ID doesn't exists this method does nothing.
func (db *Database) DeleteChat(id int64) error {
	if err := db.migrateOldChats(); err != nil {
		return fmt.Errorf("on migrating old chat's database: %w", err)
	}

	// Chat info are stored on multiple keys on DB, we must remove each one.
	sid := strconv.FormatInt(id, 10)
	if err := db.conn.SRem(context.TODO(), "chats", sid).Err(); err != nil {
		return fmt.Errorf("on removing the given chat ID from \"chats\" hash set: %w", err)
	}

	// Chat's details.
	hid := "chats:" + sid
	if err := db.conn.Del(context.TODO(), hid).Err(); err != nil {
		return fmt.Errorf("on removing chat info %q key: %w", hid, err)
	}

	// Remove also info stored on these keys.
	if err := db.conn.HDel(context.TODO(), "public-links", sid).Err(); err != nil {
		return fmt.Errorf("on removing chat's public link from \"public-links\": %w", err)
	}
	if err := db.conn.HDel(context.TODO(), "settings", sid).Err(); err != nil {
		return fmt.Errorf("on removing chat's settings from \"settings\": %w", err)
	}

	return nil
}

// ChatroomsCount returns the number of tracked chats.
func (db *Database) ChatroomsCount() (int64, error) {
	if err := db.migrateOldChats(); err != nil {
		return 0, fmt.Errorf("on migrating old chat's database: %w", err)
	}

	ret, err := db.conn.SCard(context.TODO(), "chats").Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return ret, nil
}

// ListMyChatrooms returns the list of tracked chats.
func (db *Database) ListMyChats() ([]*tb.Chat, error) {
	if err := db.migrateOldChats(); err != nil {
		return nil, fmt.Errorf("on migrating old chat's database: %w", err)
	}

	var chats []*tb.Chat

	var cursor uint64 = 0
	var err error
	var keys []string
	// ListMyChatrooms works by by deserializing the tb.Chat for each chatroom.
	for {
		keys, cursor, err = db.conn.SScan(context.TODO(), "chats", cursor, "", -1).Result()
		if errors.Is(err, redis.Nil) {
			return chats, nil
		} else if err != nil {
			return nil, fmt.Errorf("on scanning chats in \"chats\": %w", err)
		}

		for _, key := range keys {
			chat := tb.Chat{}

			id, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("on parsing chat's id: %w", err)
			}
			chat.ID = id

			hid := "chats:" + key
			title, err := db.conn.HGet(context.TODO(), hid, "title").Result()
			if err != nil {
				return nil, fmt.Errorf("on retrieving chat's title from %q key: %w", hid, err)
			}
			chat.Title = title

			chats = append(chats, &chat)
		}

		// SCAN cycle end
		if cursor == 0 {
			break
		}
	}

	return chats, nil
}

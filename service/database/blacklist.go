package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	tb "gopkg.in/telebot.v3"
)

// AddBlacklist adds the given chat to the blacklist on the DB.
//
// It first removes the chat to the tracked chats, then add the chat to the
// blacklist.
//
// Only ID and Title fields in tb.Chat are saved into the blacklist.
func (db *Database) AddBlacklist(c *tb.Chat) error {
	// First remove the chat from the tracked chats.
	if err := db.DeleteChat(c.ID); err != nil {
		return fmt.Errorf("on blacklisting the given chat: %w", err)
	}

	// Then add the given chat ID to the blacklisted chats.
	id := strconv.FormatInt(c.ID, 10)
	if err := db.conn.SAdd(context.TODO(), "blacklist", id).Err(); err != nil {
		return fmt.Errorf("on adding the given chat to \"blacklist\" hash set: %w", err)
	}

	// Save only chat's title (ID are not human-friendly).
	hid := "blacklist:" + id
	if err := db.conn.HSet(context.TODO(), hid, "title", c.Title).Err(); err != nil {
		return fmt.Errorf("on adding the given chat title to %q hash: %w", hid, err)
	}

	return nil
}

// DeleteBlacklist removes the group of the given ID from the blacklist.
//
// If the given chat ID doesn't exist this method does nothing.
func (db *Database) DeleteBlacklist(id int64) error {
	// Chat info are stored on multiple keys on DB, we must remove each one.
	sid := strconv.FormatInt(id, 10)
	if err := db.conn.SRem(context.TODO(), "blacklist", sid).Err(); err != nil {
		return fmt.Errorf("on removing the given chat ID from \"blacklist\" hash set: %w", err)
	}

	// Chat's title.
	hid := "blacklist:" + sid
	if err := db.conn.Del(context.TODO(), hid).Err(); err != nil {
		return fmt.Errorf("on removing chat info %q key: %w", hid, err)
	}

	return nil
}

// ListBlacklist returns the list of chats that are on the blacklist.
//
// The returned tb.Chat contains only ID and Title fields.
func (db *Database) ListBlacklist() ([]*tb.Chat, error) {
	var chats []*tb.Chat
	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.conn.SScan(context.TODO(), "blacklist", cursor, "", -1).Result()
		if errors.Is(err, redis.Nil) {
			return chats, nil
		} else if err != nil {
			return nil, fmt.Errorf("on scanning chats in \"blacklist\": %w", err)
		}

		for _, key := range keys {
			chat := tb.Chat{}

			id, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("on parsing chat's id: %w", err)
			}
			chat.ID = id

			hid := "blacklist:" + key
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

// GetBlacklist returns the chat corresponding to the given ID.
//
// The returned tb.Chat contains only ID and Title fields. If the given chat ID
// doesn't exist it returns an error.
func (db *Database) GetBlacklist(id int64) (*tb.Chat, error) {
	chat := &tb.Chat{ID: id}
	sid := strconv.FormatInt(id, 10)

	// Check is the given id is on the blacklist.
	if is, err := db.conn.SIsMember(context.TODO(), "blacklist", sid).Result(); err != nil {
		return nil, fmt.Errorf("on checking if given id is on \"blacklist\" set: %w", err)
	} else if !is {
		return nil, errors.New("given id is not on blacklist")
	}

	// Retrieve chat's info.
	hid := "blacklist:" + sid
	title, err := db.conn.HGet(context.TODO(), hid, "title").Result()
	if err != nil {
		return nil, fmt.Errorf("on retrieving chat's title from %q key: %w", hid, err)
	}
	chat.Title = title

	return chat, nil
}

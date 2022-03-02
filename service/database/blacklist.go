package database

import (
	"context"
	"fmt"
	"strconv"

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
// If the given chat ID doesn't exists this method does nothing.
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

package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (db *Database) migrateOldBotAdmins() error {
	oldGlobalAdmins, err := db.conn.HExists(context.TODO(), "global", "admins").Result()
	if err != nil {
		return err
	}
	if oldGlobalAdmins {
		admins, err := db.conn.HGet(context.TODO(), "global", "admins").Result()
		if err != nil {
			return fmt.Errorf("on \"HGET global admins\": %w", err)
		}

		// Migrate every item.
		for _, sID := range strings.Split(admins, ",") {
			ID, err := strconv.ParseInt(sID, 10, 64)
			if err != nil {
				continue
			}
			if err := db.AddBotAdmin(ID); err != nil {
				return fmt.Errorf("on migrating bot admin %s: %w", sID, err)
			}
		}

		// Delete old database, otherwhise the migration is done at every call.
		if err := db.conn.HDel(context.TODO(), "global", "admins").Err(); err != nil {
			return fmt.Errorf("on \"HDEL global\": %w", err)
		}
	}
	return nil
}

// IsBotAdmin returns true if the given user id is a bot admin.
func (db *Database) IsBotAdmin(id int64) (bool, error) {
	if err := db.migrateOldBotAdmins(); err != nil {
		return false, fmt.Errorf("on migrating old database: %w", err)
	}

	sid := strconv.FormatInt(id, 10)
	is, err := db.conn.SIsMember(context.TODO(), "global-admins", sid).Result()
	if err != nil {
		return false, fmt.Errorf("on SISMEMBER \"global-admins\": %w", err)
	}
	return is, nil
}

// AddBotAdmin adds the given user id as a bot admin.
func (db *Database) AddBotAdmin(id int64) error {
	sid := strconv.FormatInt(id, 10)
	if err := db.conn.SAdd(context.TODO(), "global-admins", sid).Err(); err != nil {
		return fmt.Errorf("on \"SADD global-admins\": %w", err)
	}
	return nil
}

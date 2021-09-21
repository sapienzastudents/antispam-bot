package botdatabase

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// ErrInviteLinkNotFound is returned when the invite link was not found in the database
var ErrInviteLinkNotFound = errors.New("Invite link not found")

func (db *_botDatabase) GetInviteLink(chatID int64) (string, error) {
	ret, err := db.redisconn.HGet("invitelinks", fmt.Sprint(chatID)).Result()
	if err == redis.Nil {
		return "", ErrInviteLinkNotFound
	}
	return ret, err
}

func (db *_botDatabase) SetInviteLink(chatID int64, inviteLink string) error {
	return db.redisconn.HSet("invitelinks", fmt.Sprint(chatID), inviteLink).Err()
}

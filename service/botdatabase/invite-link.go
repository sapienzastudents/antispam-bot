package botdatabase

import (
	"strconv"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// ErrInviteLinkNotFound is returned when the invite link was not found in the database
var ErrInviteLinkNotFound = errors.New("Invite link not found")

// GetInviteLink returns the cached invite link
func (db *_botDatabase) GetInviteLink(chatID int64) (string, error) {
	ret, err := db.redisconn.HGet("invitelinks", strconv.FormatInt(chatID, 10)).Result()
	if err == redis.Nil {
		return "", ErrInviteLinkNotFound
	}
	return ret, err
}

// SetInviteLink save the invite link
func (db *_botDatabase) SetInviteLink(chatID int64, inviteLink string) error {
	return db.redisconn.HSet("invitelinks", strconv.FormatInt(chatID, 10), inviteLink).Err()
}

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

const (
	ActionNone      = 0
	ActionMute      = 1
	ActionKick      = 2
	ActionBan       = 3
	ActionDeleteMsg = 4
)

var (
	ErrChatNotFound = errors.New("chat not found")
)

type ChatAdminList []int64

// IsAdmin returns true if the given user is a chat admin.
//
// Time complexity: O(n) where "n" is the number of admins in the chat.
func (list *ChatAdminList) IsAdmin(user *tb.User) bool {
	if list == nil {
		return false
	}
	for _, v := range *list {
		if v == user.ID {
			return true
		}
	}
	return false
}

// SetFromChat updates the admin list from the slice of chat members from the bot
//
// Time complexity: O(n) where "n" is the number of admins in the chat.
func (list *ChatAdminList) SetFromChat(admins []tb.ChatMember) {
	*list = ChatAdminList{}
	for _, u := range admins {
		*list = append(*list, u.User.ID)
	}
}

type BotAction struct {
	// Action is what the bot should do
	Action int `json:"action"`

	// Duration is the duration of the action (ban, for example), in seconds (if applicable). Zero means forever
	Duration uint `json:"duration"`

	// Delay is the amount of seconds after which the action should be performed. Zero means "immediately"
	Delay uint `json:"delay"`
}

type ChatSettings struct {
	// BotEnabled represent whether the bot is enabled for this chat. Enabling the bot will enable automatic actions
	// (such as antispam or CAS blacklist) and will enable some commands.
	BotEnabled bool `json:"bot_enabled"`

	// OnJoinDelete enables the deletion of "user joined" messages
	OnJoinDelete bool `json:"on_join_delete"`

	// OnLeaveDelete enables the deletion of "user left" messages
	OnLeaveDelete bool `json:"on_leave_delete"`

	// OnJoinChinese is the action that the bot should do if it detects a user with Chinese username or name joining the
	// chatroom
	OnJoinChinese BotAction `json:"on_join_chinese"`

	// OnJoinArabic is the action that the bot should do if it detects a user with Arabic username or name joining the
	// chatroom
	OnJoinArabic BotAction `json:"on_join_arabic"`

	// OnMessageChinese is the action that the bot should do if it detects a message in Chinese
	OnMessageChinese BotAction `json:"on_message_chinese"`

	// OnMessageArabic is the action that the bot should do if it detects a message in Arabic
	OnMessageArabic BotAction `json:"on_message_arabic"`

	//OnMessageSpam    BotAction `json:"on_message_spam"`

	// OnBlacklistCAS is the action that the bot should do if it detects a message from a CAS-banned user
	OnBlacklistCAS BotAction `json:"on_blacklist_cas"`

	// ChatAdmins is the list of chat admins (regardless of their permissions in the chat)
	ChatAdmins ChatAdminList `json:"chat_admins"`

	// Hidden specify if the bot is hidden in the chat list (for users)
	Hidden bool `json:"hidden"`

	// MainCategory is the general category for this chat
	MainCategory string `json:"main_category"`

	// SubCategory is the inner category for this chat
	SubCategory string `json:"sub_category"`

	// LogChannel is the channel where the bot logs its own actions. If zero, no logs will occur
	LogChannel int64 `json:"log_channel"`
}

// GetChatSettings returns the chat settings of the bot for the given chat ID.
func (db *Database) GetChatSettings(chatID int64) (ChatSettings, error) {
	// GetChatSettings deserializes the JSON with the ChatSettings structure
	// inside the "settings" HSET (the field name is the chat ID as string).
	settings := ChatSettings{}
	jsonb, err := db.conn.HGet(context.TODO(), "settings", strconv.FormatInt(chatID, 10)).Result()
	if err == redis.Nil {
		return settings, ErrChatNotFound
	} else if err != nil {
		return ChatSettings{}, err
	}

	if err = json.Unmarshal([]byte(jsonb), &settings); err != nil {
		return settings, fmt.Errorf("error decoding chat settings from JSON: %w", err)
	}
	return settings, nil
}

// SetChatSettings saves the chat settings of the bot for the given chat ID.
func (db *Database) SetChatSettings(chatID int64, settings ChatSettings) error {
	// SetChatSettings saves the settings by serializing it into a JSON, and
	// puts it in the "settings" HSET (the field name is the chat ID as string).
	jsonb, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return db.conn.HSet(context.TODO(), "settings", strconv.FormatInt(chatID, 10), jsonb).Err()
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/patrickmn/go-cache"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	ACTION_NONE       = 0
	ACTION_MUTE       = 1
	ACTION_KICK       = 2
	ACTION_BAN        = 3
	ACTION_DELETE_MSG = 4
)

type ChatAdminList []tb.ChatMember

func (list *ChatAdminList) IsAdmin(user *tb.User) bool {
	for _, v := range *list {
		if v.User.ID == user.ID {
			return true
		}
	}
	return false
}

type BotAction struct {
	// Action in effect
	Action int `json:"action"`

	// Action effect, in seconds (if applicable)
	// Zero means forever
	Duration uint `json:"duration"`

	// Action delay, zero means "immediately"
	Delay uint `json:"delay"`
}

type ChatSettings struct {
	// Whether the bot is enabled for this chat
	BotEnabled bool `json:"bot_enabled"`

	OnJoinDelete  bool `json:"on_join_delete"`
	OnLeaveDelete bool `json:"on_leave_delete"`

	OnJoinChinese BotAction `json:"on_join_chinese"`
	OnJoinArabic  BotAction `json:"on_join_arabic"`

	// Action on specific messages patterns
	OnMessageChinese BotAction `json:"on_message_chinese"`
	OnMessageArabic  BotAction `json:"on_message_arabic"`
	//OnMessageSpam    BotAction `json:"on_message_spam"`

	//OnBlacklistCAS BotAction `json:"on_blacklist_cas"`

	// Chat admins (only for cache)
	ChatAdmins ChatAdminList `json:"-"`
}

func (db *_botDatabase) GetChatSetting(chat *tb.Chat) (ChatSettings, error) {
	settingscache, found := db.cache.Get(fmt.Sprintf("chat:%d:settings", chat.ID))
	if !found {
		settings := ChatSettings{}
		jsonb, err := db.redisconn.HGet("settings", fmt.Sprintf("%d", chat.ID)).Result()
		if err == redis.Nil {
			// Settings not found, load default values
			settings = ChatSettings{
				BotEnabled:    true,
				OnJoinDelete:  false,
				OnLeaveDelete: false,
				OnJoinChinese: BotAction{
					Action:   ACTION_NONE,
					Duration: 0,
					Delay:    0,
				},
				OnJoinArabic: BotAction{
					Action:   ACTION_NONE,
					Duration: 0,
					Delay:    0,
				},
				OnMessageChinese: BotAction{
					Action:   ACTION_NONE,
					Duration: 0,
					Delay:    0,
				},
				OnMessageArabic: BotAction{
					Action:   ACTION_NONE,
					Duration: 0,
					Delay:    0,
				},
				ChatAdmins: []tb.ChatMember{},
			}
		} else if err != nil {
			return ChatSettings{}, err
		} else {
			err = json.Unmarshal([]byte(jsonb), &settings)
			if err != nil {
				return ChatSettings{}, err
			}
		}

		settings.ChatAdmins, err = b.AdminsOf(chat)
		if err != nil {
			return ChatSettings{}, err
		}

		db.cache.Set(fmt.Sprintf("chat:%d:settings", chat.ID), settings, cache.DefaultExpiration)
		return settings, nil
	} else {
		return settingscache.(ChatSettings), nil
	}
}

func (db *_botDatabase) SetChatSettings(chat *tb.Chat, settings ChatSettings) error {
	jsonb, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	err = db.redisconn.HSet("settings", fmt.Sprintf("%d", chat.ID), jsonb).Err()
	if err != nil {
		return err
	}

	db.cache.Set(fmt.Sprintf("chat:%d:settings", chat.ID), settings, cache.DefaultExpiration)
	return nil
}

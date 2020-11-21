package botdatabase

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
	"sort"
	"strings"
)

const ChatCategorySeparator = "||"

type ChatCategory string

func (c *ChatCategory) Set(s []string) {
	*c = ChatCategory(strings.Join(s, ChatCategorySeparator))
}

func (c *ChatCategory) Get() []string {
	return strings.Split(string(*c), ChatCategorySeparator)
}

type ChatCategoryTree struct {
	Chats         []*tb.Chat
	SubCategories map[string]ChatCategoryTree
}

func (c *ChatCategoryTree) GetSubCategoryList() []string {
	var ret []string
	for subcat := range c.SubCategories {
		ret = append(ret, subcat)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}

func (c *ChatCategoryTree) GetChats() []*tb.Chat {
	sort.Slice(c.Chats, func(i, j int) bool {
		return c.Chats[i].Title < c.Chats[j].Title
	})
	return c.Chats
}

func (db *_botDatabase) GetChatCategory(c *tb.Chat) (ChatCategory, error) {
	category, err := db.redisconn.HGet("chat-categories", fmt.Sprint(c.ID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return ChatCategory(category), err
}

func (db *_botDatabase) SetChatCategory(c *tb.Chat, cat ChatCategory) error {
	return db.redisconn.HSet("chat-categories", fmt.Sprint(c.ID), cat).Err()
}

func (db *_botDatabase) GetChatTree() (ChatCategoryTree, error) {
	var ret = ChatCategoryTree{}

	chatrooms, err := db.ListMyChatrooms()
	if err != nil {
		return ret, err
	}

	for _, v := range chatrooms {
		category, err := db.GetChatCategory(v)
		if err != nil {
			return ret, errors.Wrap(err, fmt.Sprintf("can't get chat category for chat %d", v.ID))
		}
		categoryLevels := category.Get()
		if category == "" {
			ret.Chats = append(ret.Chats, v)
		} else {
			if ret.SubCategories == nil {
				ret.SubCategories = map[string]ChatCategoryTree{}
			}
			if _, ok := ret.SubCategories[categoryLevels[0]]; !ok {
				ret.SubCategories[categoryLevels[0]] = ChatCategoryTree{}
			}
			var maincat = ret.SubCategories[categoryLevels[0]]

			if len(categoryLevels) == 1 {
				maincat.Chats = append(maincat.Chats, v)
			} else {
				if maincat.SubCategories == nil {
					maincat.SubCategories = map[string]ChatCategoryTree{}
				}
				if _, ok := maincat.SubCategories[categoryLevels[1]]; !ok {
					maincat.SubCategories[categoryLevels[1]] = ChatCategoryTree{}
				}
				var subcat = maincat.SubCategories[categoryLevels[1]]
				subcat.Chats = append(subcat.Chats, v)
				maincat.SubCategories[categoryLevels[1]] = subcat
			}
			ret.SubCategories[categoryLevels[0]] = maincat
		}
	}

	return ret, nil
}

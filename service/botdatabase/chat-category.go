package botdatabase

import (
	"fmt"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
	"sort"
)

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

func (db *_botDatabase) GetChatTree(b *tb.Bot) (ChatCategoryTree, error) {
	var ret = ChatCategoryTree{}

	chatrooms, err := db.ListMyChatrooms()
	if err != nil {
		return ret, err
	}

	for _, v := range chatrooms {
		settings, err := db.GetChatSetting(b, v)
		if err != nil {
			return ret, errors.Wrap(err, fmt.Sprintf("can't get chat category for chat %d", v.ID))
		}
		if settings.MainCategory == "" {
			ret.Chats = append(ret.Chats, v)
		} else {
			if ret.SubCategories == nil {
				ret.SubCategories = map[string]ChatCategoryTree{}
			}
			if _, ok := ret.SubCategories[settings.MainCategory]; !ok {
				ret.SubCategories[settings.MainCategory] = ChatCategoryTree{}
			}
			var maincat = ret.SubCategories[settings.MainCategory]

			if settings.SubCategory == "" {
				maincat.Chats = append(maincat.Chats, v)
			} else {
				if maincat.SubCategories == nil {
					maincat.SubCategories = map[string]ChatCategoryTree{}
				}
				if _, ok := maincat.SubCategories[settings.SubCategory]; !ok {
					maincat.SubCategories[settings.SubCategory] = ChatCategoryTree{}
				}
				var subcat = maincat.SubCategories[settings.SubCategory]
				subcat.Chats = append(subcat.Chats, v)
				maincat.SubCategories[settings.SubCategory] = subcat
			}
			ret.SubCategories[settings.MainCategory] = maincat
		}
	}

	return ret, nil
}

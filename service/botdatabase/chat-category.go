package botdatabase

import (
	"fmt"
	"sort"

	tb "gopkg.in/tucnak/telebot.v3"
)

// ChatCategoryTree is a node of the chat tree. Each node can contain some chats and/or some sub categories
type ChatCategoryTree struct {
	Chats         []*tb.Chat
	SubCategories map[string]ChatCategoryTree
}

// GetSubCategoryList returns an ordered list of sub categories in this node.
//
// Time complexity: O(n + n*log(n)) where "n" is the number of chats.
func (c *ChatCategoryTree) GetSubCategoryList() []string {
	var ret []string
	for subcat := range c.SubCategories {
		ret = append(ret, subcat)
	}
	sort.Strings(ret)
	return ret
}

// GetChats returns an ordered list of chats in this node.
//
// Time complexity: O(n*log(n)) where "n" is the number of chats.
func (c *ChatCategoryTree) GetChats() []*tb.Chat {
	sort.Slice(c.Chats, func(i, j int) bool {
		return c.Chats[i].Title < c.Chats[j].Title
	})
	return c.Chats
}

// GetChatTree returns the chat tree (categories).
//
// It builds the chat tree by calling GetChatSettings for each chatroom, and
// creating a second level in the tree using sub categories. This means that
// this function can return only tree with two levels, for now.
//
//
// Time complexity: O(n) where "n" is the number of chatroom where the bot is.
func (db *_botDatabase) GetChatTree() (ChatCategoryTree, error) {
	ret := ChatCategoryTree{}

	// Get the flat list of chatrooms where the bot is.
	chatrooms, err := db.ListMyChatrooms()
	if err != nil {
		return ret, err
	}

	// Rationale: for each chatroom we retrieve the chatroom settings (as both categories are inside the ChatSettings)
	// and we divide three cases
	// * the main category is empty (then they're on the root)
	// * the main category exists but the subcategory is empty (they're on the main category node of the tree)
	// * both main and sub categories exists (they're a leaf in the tree)
	//
	// Chats in the root might be "lost" or "waiting for category", so they might be hidden or treated in a special way
	for _, v := range chatrooms {
		settings, err := db.GetChatSettings(v.ID)
		if err != nil {
			return ret, fmt.Errorf("failed to get chat category for chat %d: %w", v.ID, err)
		}

		if settings.MainCategory == "" {
			// Chat is on the root node
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
				// Chat is on the intermediate node
				maincat.Chats = append(maincat.Chats, v)
			} else {
				// Chat is on the leaf
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

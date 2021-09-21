package tbot

import (
	"github.com/patrickmn/go-cache"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"math/rand"
)

// State represent a state for telegram callback
type State struct {
	Chat *tb.Chat

	// Used in category settings
	Category      string
	SubCategory   string
	SubCategories botdatabase.ChatCategoryTree
}

// CallbackStateful returns a callback function with the state injected
func (bot *telegramBot) CallbackStateful(fn func(callback *tb.Callback, state State)) func(callback *tb.Callback) {
	return func(callback *tb.Callback) {
		state, ok := bot.statemgmt.Get(callback.Data)
		if !ok {
			_ = bot.telebot.Respond(callback)
		} else {
			bot.statemgmt.Delete(callback.Data)
			if statemap, ok := state.(State); ok {
				fn(callback, statemap)
			} else {
				_ = bot.telebot.Respond(callback)
			}
		}
	}
}

func (bot *telegramBot) newCallbackState(state State) string {
	stateID := randStringRunes(20)
	bot.statemgmt.Set(stateID, state, cache.DefaultExpiration)
	return stateID
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

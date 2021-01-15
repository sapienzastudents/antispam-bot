package main

import (
	"math/rand"
	"time"

	"github.com/patrickmn/go-cache"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

var statemgmt = cache.New(1*time.Minute, 1*time.Minute)

// State represent a state for telegram callback
type State struct {
	Chat *tb.Chat

	// Used in category settings
	Category      string
	SubCategory   string
	SubCategories botdatabase.ChatCategoryTree
}

// CallbackStateful returns a callback function with the state injected
func CallbackStateful(fn func(callback *tb.Callback, state State)) func(callback *tb.Callback) {
	return func(callback *tb.Callback) {
		state, ok := statemgmt.Get(callback.Data)
		if !ok {
			_ = b.Respond(callback)
		} else {
			statemgmt.Delete(callback.Data)
			if statemap, ok := state.(State); ok {
				fn(callback, statemap)
			} else {
				_ = b.Respond(callback)
			}
		}
	}
}

func newCallbackState(state State) string {
	stateID := randStringRunes(20)
	statemgmt.Set(stateID, state, cache.DefaultExpiration)
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

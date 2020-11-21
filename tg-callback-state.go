package main

import (
	"github.com/patrickmn/go-cache"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"math/rand"
	"time"
)

var statemgmt = cache.New(1*time.Minute, 1*time.Minute)

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

func NewCallbackState(state State) string {
	stateId := RandStringRunes(20)
	statemgmt.Set(stateId, state, cache.DefaultExpiration)
	return stateId
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

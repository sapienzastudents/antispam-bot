package tbot

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	tb "gopkg.in/tucnak/telebot.v2"
)

// State represent a state for telegram callback
type State struct {
	ChatToEdit        *tb.Chat
	AddGlobalCategory bool
	AddSubCategory    bool

	bot             *telegramBot
	user            *tb.User
	chatWithTheUser *tb.Chat
}

func (s *State) Save() {
	s.bot.statemgmt.Set(fmt.Sprintf("%d %d", s.user.ID, s.chatWithTheUser.ID), *s, cache.DefaultExpiration)
}

// Create a new empty state for the user
func (bot *telegramBot) newState(user *tb.User, chatWithTheUser *tb.Chat) State {
	return State{
		user:            user,
		chatWithTheUser: chatWithTheUser,
		bot:             bot,
	}
}

// Get the current state
func (bot *telegramBot) getStateFor(user *tb.User, chat *tb.Chat) State {
	state, ok := bot.statemgmt.Get(fmt.Sprintf("%d %d", user.ID, chat.ID))
	if !ok {
		state = bot.newState(user, chat)
	} else if _, ok := state.(State); !ok {
		state = bot.newState(user, chat)
	}
	return state.(State)
}

// handleAdminCallbackStateful handles an admin restricted callback returning the state
func (bot *telegramBot) handleAdminCallbackStateful(endpoint interface{}, fn func(callback *tb.Callback, state State)) {
	bot.telebot.Handle(endpoint, func(callback *tb.Callback) {
		state := bot.getStateFor(callback.Sender, callback.Message.Chat)

		// Check if the user is authorized for this callback
		settings, err := bot.getChatSettings(state.ChatToEdit)
		if err != nil {
			bot.logger.WithError(err).Error("error getting chat settings in handleAdminCallbackStateful")
			return
		}

		isGlobalAdmin, err := bot.db.IsGlobalAdmin(callback.Sender.ID)
		if err != nil {
			bot.logger.WithError(err).Error("can't check if the user is a global admin")
			return
		}

		if settings.ChatAdmins.IsAdmin(callback.Sender) || isGlobalAdmin {
			// User authorized
			fn(callback, state)
			return
		}

		// User not authorized
		_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
			Text:      "Not authorized",
			ShowAlert: false,
		})
	})
}

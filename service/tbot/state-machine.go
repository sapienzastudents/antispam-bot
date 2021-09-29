package tbot

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Rationale: bot requests are stateless. However, we need to maintain some state across bot commands (especially in
// settings commands). This file contains the implementation for a state machine that stores some state and some infos
// about settings.

// State represent a state of the bot state machine
type State struct {
	// ChatToEdit is the chat the user is editing
	ChatToEdit *tb.Chat

	// AddGlobalCategory is a flag indicating that the next text message is the name of the global category
	AddGlobalCategory bool

	// AddSubCategory is a flag indicating that the next message is the name of the sub category
	AddSubCategory bool

	bot             *telegramBot
	user            *tb.User
	chatWithTheUser *tb.Chat
}

// Save serialize and save the current state
//
// Time complexity: O(1)
func (s *State) Save() {
	s.bot.statemgmt.Set(fmt.Sprintf("%d %d", s.user.ID, s.chatWithTheUser.ID), *s, cache.DefaultExpiration)
}

// newState creates a new empty state for the user
//
// Time complexity: O(1)
func (bot *telegramBot) newState(user *tb.User, chatWithTheUser *tb.Chat) State {
	return State{
		user:            user,
		chatWithTheUser: chatWithTheUser,
		bot:             bot,
	}
}

// getStateFor returns the current state for the user
//
// Time complexity: O(1)
func (bot *telegramBot) getStateFor(user *tb.User, chat *tb.Chat) State {
	state, ok := bot.statemgmt.Get(fmt.Sprintf("%d %d", user.ID, chat.ID))
	if !ok {
		state = bot.newState(user, chat)
	} else if _, ok := state.(State); !ok {
		state = bot.newState(user, chat)
	}
	return state.(State)
}

// handleAdminCallbackStateful register a handler for an admin action callback, injecting the user state. The callback
// restrict the action callback to a chat admin or a global admin (in other words, the callback sender must be an admin)
func (bot *telegramBot) handleAdminCallbackStateful(endpoint interface{}, fn func(callback *tb.Callback, state State)) {
	bot.telebot.Handle(endpoint, func(callback *tb.Callback) {
		state := bot.getStateFor(callback.Sender, callback.Message.Chat)

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
			// User authorized, call the registered function
			fn(callback, state)
		} else {
			_ = bot.telebot.Respond(callback, &tb.CallbackResponse{
				Text:      "Not authorized",
				ShowAlert: false,
			})
		}
	})
}

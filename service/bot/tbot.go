// Package bot implements the core of the Telegram bot. It handles all telegram commands and events. Requires a
// database to store all chat settings and cache. Optionally, it can use the CAS database and it can update a website
// sitting in a git repo (via an SSH key) for the group list feature.
//
// The bot handles some commands from anyone, and some others from chat admins only. These are registered in
// ListenAndServe() function. They are always wrapped with some utility functions like the one for metrics, and
// refreshDBInfo for keeping track of chats (due to the fact that bots are not allowed to "know" in which chat they're
// in from telegram APIs).
//
// Global admins are always allowed to talk to the bot and do things like fixing settings.
//
// Nearly all commands are stateless. Settings however is stateful (meaning that we need to keep track between
// subsequents actions). A simple state loader and saver has been written in state-machine.go, even if it's not a
// complete FSM.
//
// This package is the result of the refactoring of an old antispam Telegram bot. We're in process of cleaning up the
// code and fix some behaviors.
package bot

import (
	"net/http"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/cas"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/database"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/i18n"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

// TelegramBot is the interface for implementations of this kind of bots
type TelegramBot interface {
	// MetricsHandler returns a HTTP handler for exposing metrics
	MetricsHandler() http.Handler

	// ListenAndServe starts the bot
	ListenAndServe() error

	// Close ends the bot
	Close() error
}

type telegramBot struct {
	// logger is a logrus instance for structured logging
	logger logrus.FieldLogger

	// db is the main database instance, use this for settings, cache and other things that needs to survive between
	// reboots
	db *database.Database

	// cas is the CAS database interface, if any
	cas cas.CAS

	// Bundle is the Bundle instance to get localized strings.
	bundle *i18n.Bundle

	// gitTemporaryDir is the directory where the bot will temporarly check out the website repository in order to
	// modify the content of the links page
	gitTemporaryDir string

	// gitSSHKey is the path to the SSH key used for authenticate against the repo of the website
	gitSSHKey string

	// gitSSHKeyPassphrase is the SSH key passphrase (see gitSSHKey)
	gitSSHKeyPassphrase string

	// telebot is an instance of the telebot library
	telebot *tb.Bot

	// promreg is a prometheus registry for metrics
	promreg *prometheus.Registry

	// statemgmt is a in-memory key-value store for the bot state machine. See state-machine.go for details
	statemgmt *cache.Cache

	// messageProcessedTotal is the counter of total processed messages
	messageProcessedTotal prometheus.Counter

	// backgroundRefreshElapsed is the time elapsed for background cache refresh
	backgroundRefreshElapsed prometheus.Gauge

	// groupUserCount is the number of users per chat
	groupUserCount *prometheus.GaugeVec

	// groupMessagesCount message count per chat
	groupMessagesCount *prometheus.CounterVec

	// userMessageCount is the message count per user
	userMessageCount *prometheus.CounterVec

	/*casDatabaseDownloadTime  prometheus.Gauge
	casDatabaseSize          prometheus.Gauge*/

	// casDatabaseMatch is the total number of matches for CAS
	casDatabaseMatch prometheus.Counter

	// botCommandsRequestsTotal is the number of requests per command
	botCommandsRequestsTotal *prometheus.CounterVec

	// botReplyLatency is the latency of reply of the bot
	botReplyLatency prometheus.Histogram
}

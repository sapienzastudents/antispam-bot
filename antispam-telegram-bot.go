package main

import (
	"fmt"
	"github.com/op/go-logging"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"time"
)

var APP_VERSION = "dev"

var b *tb.Bot = nil
var logger *logging.Logger = nil
var botdb BOTDatabase = nil

func main() {
	var err error

	// Logging init
	logger = logging.MustGetLogger("goinfostudapi")

	// To add logging to file, edit this section and implement a system based on environmental variables
	backend1Leveled := logging.AddModuleLevel(logging.NewLogBackend(os.Stdout, "", 0))
	backend1Leveled.SetLevel(logging.DEBUG, "")
	logger.SetBackend(backend1Leveled)

	logger.Info("Initializing...")
	// Initial checks
	if os.Getenv("BOT_TOKEN") == "" {
		logger.Fatal("BOT_TOKEN environment variable not set - exiting")
		return
	}
	if os.Getenv("REDIS_URL") == "" {
		logger.Fatal("REDIS_URL environment variable not set - exiting")
		return
	}

	// Initialize Redis database
	botdb, err = NewBotDatabase()
	if err != nil {
		logger.Fatalf("Unable to initialize the database, exiting: %s", err.Error())
		return
	}

	// Initialize bot library
	b, err = tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		logger.Fatal(err)
		return
	}

	// Registering internal/utils handlers
	b.Handle(tb.OnVoice, HandlerWrapper(onAnyMessage))
	b.Handle(tb.OnVideo, HandlerWrapper(onAnyMessage))
	b.Handle(tb.OnEdited, HandlerWrapper(onAnyMessage))
	b.Handle(tb.OnDocument, HandlerWrapper(onAnyMessage))
	b.Handle(tb.OnAudio, HandlerWrapper(onAnyMessage))
	b.Handle(tb.OnPhoto, HandlerWrapper(onAnyMessage))
	b.Handle(tb.OnText, HandlerWrapper(onAnyMessage))

	b.Handle("/id", HandlerWrapper(func(m *tb.Message, _ ChatSettings) {
		if m.Private() {
			_, _ = b.Send(m.Chat, fmt.Sprint(m.Sender.ID))
		}
	}))

	b.Handle("/version", HandlerWrapper(func(m *tb.Message, _ ChatSettings) {
		if m.Private() {
			msg := fmt.Sprintf("Version %s", APP_VERSION)
			_, _ = b.Send(m.Chat, msg)
		}
	}))

	// Register commands
	b.Handle("/help", HandlerWrapper(onHelp))
	b.Handle("/start", HandlerWrapper(onHelp))
	//b.Handle("/unmute", onUnMuteRequest)

	// Chat-admin commands
	b.Handle("/settings", HandlerWrapper(onSettings))

	// Global-administrative commands
	b.Handle("/mychatrooms", HandlerWrapper(onMyChatroomRequest))

	// Register events
	b.Handle(tb.OnUserJoined, HandlerWrapper(onUserJoined))
	b.Handle(tb.OnUserLeft, HandlerWrapper(onUserLeft))

	logger.Info("Init ok, starting bot")

	// Cache updater
	go func() {
		time.Sleep(5 * time.Minute)

		t := time.NewTicker(5 * time.Minute)
		for {
			<-t.C
			err := botdb.DoCacheUpdate()
			if err != nil {
				logger.Critical(err)
			}
		}
	}()

	// Let's go!
	b.Start()
}

// Wrapper to each call
func HandlerWrapper(actionHandler func(*tb.Message, ChatSettings)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			err := botdb.UpdateMyChatroomList(m.Chat)
			if err != nil {
				logger.Critical("Cannot update my chatroom list:", err)
				return
			}

			settings, err := botdb.GetChatSetting(m.Chat)
			if err != nil {
				logger.Critical("Cannot get chat settings:", err)
			} else if !settings.BotEnabled {
				logger.Debugf("Bot not enabled for chat %d %s", m.Chat.ID, m.Chat.Title)
			} else {
				actionHandler(m, settings)
			}
		} else {
			actionHandler(m, ChatSettings{})
		}
	}
}

func CallbackWrapper(fn func(*tb.Callback, ChatSettings), onlyadmins bool) func(*tb.Callback) {
	return func(callback *tb.Callback) {
		settings, err := botdb.GetChatSetting(callback.Message.Chat)
		if err != nil {
			logger.Critical("Cannot get chat settings:", err)
		} else if onlyadmins && !settings.ChatAdmins.IsAdmin(callback.Sender) {
			logger.Critical("Non-admin is using a callback from the admin:", callback.Sender)
			b.Respond(callback, &tb.CallbackResponse{
				Text:      "Not authorized",
				ShowAlert: true,
			})
		} else {
			fn(callback, settings)
		}
	}
}

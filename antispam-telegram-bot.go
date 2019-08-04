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
	b.Handle(tb.OnText, func(m *tb.Message) {
		// Note: this will not scale very well - keep an eye on it
		if !m.Private() {
			if b, err := botdb.IsBotEnabled(m.Chat); !b || err != nil {
				logger.Debugf("Bot not enabled for chat %d %s", m.Chat.ID, m.Chat.Title)
				return
			}

			botdb.UpdateMyChatroomList(m.Chat)

			// Launch spam detection algorithms
			if chineseChars(m.Text) > 0.5 || arabicChars(m.Text) > 0.5 {
				actionDelete(m)
				// Or we can mute it (TODO: leave it as an option)
				//muteUser(m.Chat, m.Sender, m)
			}
		}
	})

	b.Handle("/id", func(m *tb.Message) {
		if m.Private() {
			_, _ = b.Send(m.Chat, fmt.Sprint(m.Sender.ID))
		} else {
			if b, err := botdb.IsBotEnabled(m.Chat); !b || err != nil {
				return
			}
			botdb.UpdateMyChatroomList(m.Chat)
		}
	})

	b.Handle("/version", func(m *tb.Message) {
		if m.Private() {
			msg := fmt.Sprintf("Version %s", APP_VERSION)
			_, _ = b.Send(m.Chat, msg)
		} else {
			if b, err := botdb.IsBotEnabled(m.Chat); !b || err != nil {
				return
			}
			botdb.UpdateMyChatroomList(m.Chat)
		}
	})

	// Register commands
	b.Handle("/help", onHelp)
	b.Handle("/start", onHelp)
	//b.Handle("/unmute", onUnMuteRequest)

	// Global-administrative commands
	b.Handle("/mychatrooms", onMyChatroomRequest)

	// Register events
	b.Handle(tb.OnUserJoined, onUserJoined)
	b.Handle(tb.OnUserLeft, onUserLeft)

	logger.Info("Init ok, starting bot")
	go adminListUpdater()
	// Let's go!
	b.Start()
}

func adminListUpdater() {
	// TODO: completare questa parte prima di metterlo in produzione
	/*
		// The first update is in 5 seconds
		time.Sleep(5 * time.Second)

		// Then, every 5 minutes
		t := time.NewTicker(5 * time.Minute)
		for {
			<-t.C
			startms := time.Now()
			var cursor uint64 = 0
			var keys []string
			var err error
			for {
				keys, cursor, err = redisconn.Scan(cursor, "chat.*", 10).Result()
				if err != nil {
					logger.Criticalf("Cannot scan for admins, redis error: %s", err.Error())
					break
				} else {

					// TODO: scan for admins and cache it

					if cursor == 0 {
						// The end of scan
						break
					}
				}
			}

			logger.Infof("Chat admin scan done in %.2f seconds", time.Now().Sub(startms).Seconds())
		}

		// This is not a very clean way to finish - however it works, and hopefully the bot is not rebooted as it always works®
	*/
}

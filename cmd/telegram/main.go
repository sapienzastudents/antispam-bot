package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/cas"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/tbot"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// AppVersion is the app version injected by the compiler
var AppVersion = "dev"

// BuildDate is the app build date injected by the compiler
var BuildDate = "n/a"

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize environment
	rand.Seed(time.Now().UnixNano())
	_ = godotenv.Load()

	// Create buffered channel for signals
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM)

	// Logging init
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(os.Stdout)
	logrusLogger.SetLevel(logrus.DebugLevel) // TODO
	// TODO: logrusLogger.SetFormatter(&logrus.JSONFormatter{})

	// Setting the hostname if available
	hostname, _ := os.Hostname()
	logger := logrusLogger.WithFields(logrus.Fields{
		"hostname": hostname,
	})

	logger.Info("Antispam Telegram Bot, version ", AppVersion, " build ", BuildDate)

	// Initalizing redis connection
	logger.Info("Initializing redis DB connection")
	redisOptions, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return errors.Wrap(err, "unable to parse REDIS_URL variable")
	}
	redisDb := redis.NewClient(redisOptions)
	err = redisDb.Ping(context.TODO()).Err()
	if err != nil {
		return errors.Wrap(err, "unable to connect to Redis server")
	}
	botdb, err := botdatabase.New(redisDb)
	if err != nil {
		return errors.Wrap(err, "error creating DB connection")
	}

	// Initializing CAS database
	casDB, err := cas.New(os.Getenv("DISABLE_CAS") == "", logger, nil)
	if err != nil {
		return errors.Wrap(err, "error creating CAS")
	}

	// Initialize Telegram
	logger.Info("Initializing Telegram bot connection")
	bot, err := tbot.New(tbot.Options{
		Logger:              logrusLogger,
		Database:            botdb,
		Token:               os.Getenv("BOT_TOKEN"),
		CAS:                 casDB,
		GitTemporaryDir:     os.Getenv("GIT_TEMP_DIR"),
		GitSSHKeyFile:       os.Getenv("GIT_SSH_KEY"),
		GitSSHKeyPassphrase: os.Getenv("GIT_SSH_KEY_PASS"),
	})
	if err != nil {
		return errors.Wrap(err, "error creating bot")
	}

	_ = os.Setenv("GIT_SSH_KEY_PASS", "")
	_ = os.Setenv("GIT_SSH_KEY", "")

	go func() {
		// Temporary HTTP Server for metrics
		http.Handle("/metrics", bot.MetricsHandler())
		_ = http.ListenAndServe(":3000", nil)
	}()

	go func() {
		<-sigchan
		_ = bot.Close()
	}()

	err = bot.ListenAndServe()
	return errors.Wrap(err, "error starting bot")
}

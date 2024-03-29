package bot

import (
	"errors"
	"time"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/cas"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/database"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/i18n"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

type Options struct {
	// Logger is a logrus logger for program errors and debug infos. Required
	Logger logrus.FieldLogger

	// Database is needed for chat cache and settings. Required
	Database *database.Database

	// Token is the Telegram bot token, from BotFather. Required
	Token string

	// CAS is the CAS database instance
	CAS cas.CAS

	// Bundle is the Bundle instance to get localized strings. Required.
	Bundle *i18n.Bundle

	// GitTemporaryDir is a temporary directory for git operations
	GitTemporaryDir string

	// GitSSHKeyFile is the SSH key file path for git push/pull
	GitSSHKeyFile string

	// GitSSHKeyPassphrase is the SSH key passphrase
	GitSSHKeyPassphrase string

	// LongPollerTimeout is the timeout for long polling. Default: 10s
	LongPollerTimeout time.Duration
}

// New returns a new TelegramBot compliant instance.
//
// It returns an error when either one of the required parameters are not set,
// or if there is an error talking with Telegram servers.
func New(opts Options) (TelegramBot, error) {
	if opts.Logger == nil {
		return nil, errors.New("logger not present")
	}
	if opts.Database == nil {
		return nil, errors.New("database not present")
	}
	if opts.Token == "" {
		return nil, errors.New("telegram bot token not specified")
	}
	if opts.Bundle == nil {
		return nil, errors.New("bundle not specified")
	}

	if opts.LongPollerTimeout == 0 {
		opts.LongPollerTimeout = 10 * time.Second
	}

	// Initialize bot library
	telebot, err := tb.NewBot(tb.Settings{
		Token:  opts.Token,
		Poller: &tb.LongPoller{Timeout: opts.LongPollerTimeout},
	})
	if err != nil {
		return nil, err
	}

	t := telegramBot{
		logger:              opts.Logger,
		db:                  opts.Database,
		cas:                 opts.CAS,
		bundle:              opts.Bundle,
		gitTemporaryDir:     opts.GitTemporaryDir,
		gitSSHKey:           opts.GitSSHKeyFile,
		gitSSHKeyPassphrase: opts.GitSSHKeyPassphrase,
		telebot:             telebot,
	}

	t.statemgmt = cache.New(60*time.Minute, 60*time.Minute)

	// Initialize metrics
	t.promreg = prometheus.NewRegistry()

	// General
	t.messageProcessedTotal = promauto.With(t.promreg).NewCounter(prometheus.CounterOpts{
		Name: "bot_message_processed_total",
		Help: "The number of total messages processed by the bot",
	})
	t.backgroundRefreshElapsed = promauto.With(t.promreg).NewGauge(prometheus.GaugeOpts{
		Name: "bot_background_refresh_elapsed",
		Help: "The elapsed time for backgroud refresh of all chatrooms",
	})

	// Per chatroom
	_ = promauto.With(t.promreg).NewGaugeFunc(prometheus.GaugeOpts{
		Name: "bot_students_groups",
		Help: "The total number of student groups in the index",
	}, func() float64 {
		ret, err := t.db.ChatroomsCount()
		if err != nil {
			t.logger.WithError(err).Error("can't get chatrooms count")
			return 0
		}
		return float64(ret)
	})
	t.groupUserCount = promauto.With(t.promreg).NewGaugeVec(prometheus.GaugeOpts{
		Name: "bot_group_user_count",
		Help: "The total number of users per group",
	}, []string{"chatid", "chatname"})
	t.groupMessagesCount = promauto.With(t.promreg).NewCounterVec(prometheus.CounterOpts{
		Name: "bot_group_messages_count",
		Help: "The total number of messages per group",
	}, []string{"chatid", "chatname"})
	t.userMessageCount = promauto.With(t.promreg).NewCounterVec(prometheus.CounterOpts{
		Name: "bot_user_messages_count",
		Help: "The total number of messages per user",
	}, []string{"userid", "username"})

	// CAS database
	/*t.casDatabaseDownloadTime = promauto.With(t.promreg).NewGauge(prometheus.GaugeOpts{
		Name: "cas_database_download_time",
		Help: "The time elapsed for downloading the CAS database",
	})
	t.casDatabaseSize = promauto.With(t.promreg).NewGauge(prometheus.GaugeOpts{
		Name: "cas_database_size",
		Help: "The number of items in the CAS database",
	})*/
	t.casDatabaseMatch = promauto.With(t.promreg).NewCounter(prometheus.CounterOpts{
		Name: "cas_database_match",
		Help: "The number of users in the CAS database matched",
	})

	// Bot commands
	t.botCommandsRequestsTotal = promauto.With(t.promreg).NewCounterVec(prometheus.CounterOpts{
		Name: "bot_commands_requests_total",
		Help: "The number of requests for bot commands",
	}, []string{"command"})
	t.botReplyLatency = promauto.With(t.promreg).NewHistogram(prometheus.HistogramOpts{
		Name:    "bot_reply_latency",
		Help:    "The latency in the bot response",
		Buckets: prometheus.ExponentialBuckets(5, 4, 6),
	})

	return &t, nil
}

package tbot

import (
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/cas"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
)

type TelegramBot interface {
	MetricsHandler() http.Handler
	ListenAndServe() error
	Close() error
}

type telegramBot struct {
	logger logrus.FieldLogger
	db     botdatabase.Database
	cas    cas.CAS

	telebot       *tb.Bot
	promreg       *prometheus.Registry
	statemgmt     *cache.Cache
	globaleditcat *cache.Cache

	// Metrics
	messageProcessedTotal    prometheus.Counter
	backgroundRefreshElapsed prometheus.Gauge
	groupUserCount           *prometheus.GaugeVec
	groupMessagesCount       *prometheus.CounterVec
	userMessageCount         *prometheus.CounterVec
	/*casDatabaseDownloadTime  prometheus.Gauge
	casDatabaseSize          prometheus.Gauge
	casDatabaseMatch         prometheus.Counter*/
	botCommandsRequestsTotal *prometheus.CounterVec
	botReplyLatency          prometheus.Histogram
}

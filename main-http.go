package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	// General
	messageProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bot_message_processed_total",
		Help: "The number of total messages processed by the bot",
	})
	backgroundRefreshElapsed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bot_background_refresh_elapsed",
		Help: "The elapsed time for backgroud refresh of all chatrooms",
	})

	// Per chatroom
	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "bot_students_groups",
		Help: "The total number of student groups in the index",
	}, func() float64 {
		if botdb != nil {
			ret, err := botdb.ChatroomsCount()
			if err != nil {
				logger.WithError(err).Error("can't get chatrooms count")
				return 0
			}
			return float64(ret)
		}
		return 0
	})
	groupUserCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bot_group_user_count",
		Help: "The total number of users per group",
	}, []string{"chatid", "chatname"})
	groupMessagesCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bot_group_messages_count",
		Help: "The total number of messages per group",
	}, []string{"chatid", "chatname", "hour"})
	userMessageCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bot_user_messages_count",
		Help: "The total number of messages per user",
	}, []string{"userid", "username"})

	// CAS database
	casDatabaseDownloadTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cas_database_download_time",
		Help: "The time elapsed for downloading the CAS database",
	})
	casDatabaseSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cas_database_size",
		Help: "The number of items in the CAS database",
	})
	casDatabaseMatch = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cas_database_match",
		Help: "The number of users in the CAS database matched",
	})

	// Bot commands
	botCommandsRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bot_commands_requests_total",
		Help: "The number of requests for bot commands",
	}, []string{"command"})
	botReplyLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "bot_reply_latency",
		Help:    "The latency in the bot response",
		Buckets: prometheus.ExponentialBuckets(5, 4, 6),
	})
)

func mainHTTP() {
	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(":3000", nil)
}

func metrics(fn func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		startms := time.Now()
		fn(m)
		messageProcessedTotal.Inc()
		if !m.Private() {
			groupMessagesCount.WithLabelValues(fmt.Sprint(m.Chat.ID), m.Chat.Title, fmt.Sprint(time.Now().UTC().Hour())).Inc()
			var userName string
			if m.Sender.Username == "" {
				userName = strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
			} else {
				userName = "@" + m.Sender.Username
			}
			userMessageCount.WithLabelValues(fmt.Sprint(m.Sender.ID), userName).Inc()
		}
		botReplyLatency.Observe(float64(time.Since(startms) / time.Millisecond))
	}
}

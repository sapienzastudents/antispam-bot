package tbot

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MetricsHandler returns a HTTP handler for exposing metrics
func (bot *telegramBot) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(bot.promreg, promhttp.HandlerOpts{
		Registry: bot.promreg,
		Timeout:  30 * time.Second,
	})
}

// metrics collects metrics for commands, users and chats (reply latency, message counts, and more)
func (bot *telegramBot) metrics(fn func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		startms := time.Now()
		fn(m)
		bot.messageProcessedTotal.Inc()
		if !m.Private() {
			bot.groupMessagesCount.WithLabelValues(strconv.FormatInt(m.Chat.ID, 10), m.Chat.Title).Inc()
			var userName string
			if m.Sender.Username == "" {
				userName = strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
			} else {
				userName = "@" + m.Sender.Username
			}
			bot.userMessageCount.WithLabelValues(strconv.FormatInt(int64(m.Sender.ID), 10), userName).Inc()
		}
		bot.botReplyLatency.Observe(float64(time.Since(startms) / time.Millisecond))
	}
}

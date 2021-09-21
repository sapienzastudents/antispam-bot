package tbot

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"strings"
	"time"
)

func (bot *telegramBot) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(bot.promreg, promhttp.HandlerOpts{
		Registry: bot.promreg,
		Timeout:  30 * time.Second,
	})
}

func (bot *telegramBot) metrics(fn func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		startms := time.Now()
		fn(m)
		bot.messageProcessedTotal.Inc()
		if !m.Private() {
			bot.groupMessagesCount.WithLabelValues(fmt.Sprint(m.Chat.ID), m.Chat.Title).Inc()
			var userName string
			if m.Sender.Username == "" {
				userName = strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
			} else {
				userName = "@" + m.Sender.Username
			}
			bot.userMessageCount.WithLabelValues(fmt.Sprint(m.Sender.ID), userName).Inc()
		}
		bot.botReplyLatency.Observe(float64(time.Since(startms) / time.Millisecond))
	}
}

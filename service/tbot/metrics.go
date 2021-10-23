package tbot

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	tb "gopkg.in/tucnak/telebot.v3"
)

// MetricsHandler returns a HTTP handler for exposing metrics
func (bot *telegramBot) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(bot.promreg, promhttp.HandlerOpts{
		Registry: bot.promreg,
		Timeout:  30 * time.Second,
	})
}

// metrics returns an HandlerFunc suited to be passed on bot. It wraps the given
// handler and collects metrics for commands, users and chats (reply latency,
// message counts and more).
func (bot *telegramBot) metrics(fn tb.HandlerFunc) tb.HandlerFunc {
	return func(ctx tb.Context) error {
		startms := time.Now()
		fn(ctx)
		bot.messageProcessedTotal.Inc()

		msg := ctx.Message()
		if msg == nil {
			return nil // Skip metrics for non-message updates.
		}
		if !msg.Private() {
			bot.groupMessagesCount.WithLabelValues(strconv.FormatInt(msg.Chat.ID, 10), msg.Chat.Title).Inc()
			var userName string
			if msg.Sender.Username == "" {
				userName = strings.TrimSpace(msg.Sender.FirstName + " " + msg.Sender.LastName)
			} else {
				userName = "@" + msg.Sender.Username
			}
			bot.userMessageCount.WithLabelValues(strconv.FormatInt(msg.Sender.ID, 10), userName).Inc()
		}
		bot.botReplyLatency.Observe(float64(time.Since(startms) / time.Millisecond))
		return nil
	}
}

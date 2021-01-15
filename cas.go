package main

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/levigross/grequests"
)

/*
 * This code loads the CAS Database (Combot Anti Spam)
 * You can find more info here: https://combot.org/cas/
 */

type casDB map[int]int8

var casdb = casDB{}

// This function downloads and replace current in-memory CAS database
// and returns the list of added and removed items, or nil,nil in case
// of error
func loadCAS() (casDB, casDB) {
	startms := time.Now()

	resp, err := grequests.Get("https://api.cas.chat/export.csv", &grequests.RequestOptions{
		UserAgent:      "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0",
		RequestTimeout: 1 * time.Minute,
	})
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			casDatabaseDownloadTime.Set(float64(time.Since(startms) / time.Millisecond))
			logger.WithError(err).Error("CAS database download error: timeout")
		} else {
			logger.WithError(err).Error("CAS database download error")
		}
		return nil, nil
	}

	if strings.Contains(resp.String(), "cloudflare") {
		logger.Error("CAS database download error: CloudFlare limited")
		logger.Debug(resp.String())
		return nil, nil
	}

	casDatabaseDownloadTime.Set(float64(time.Since(startms) / time.Millisecond))

	var newcas = casDB{}
	var added = casDB{}
	var deleted = casDB{}

	// Pre-load deleted array
	for id := range casdb {
		deleted[id] = 1
	}

	// Cycle new CAS database
	for _, row := range strings.Split(resp.String(), "\n") {
		if row == "" {
			continue
		}

		uid, err := strconv.Atoi(row)
		if err != nil {
			logger.WithError(err).Error("Cannot convert ID to Telegram UID")
		} else {
			// Check if it's new
			_, found := casdb[uid]
			if !found {
				added[uid] = 1
			}

			newcas[uid] = 1

			// Entry exists in the new version of the DB, so delete it from the "pending delete" list
			delete(deleted, uid)
		}
	}

	if len(newcas) > 0 {
		casdb = newcas
		casDatabaseSize.Set(float64(len(casdb)))
		logger.Debugf("CAS Database updated - items:%d, added:%d, deleted:%d", len(casdb), len(added), len(deleted))
		return added, deleted
	}
	logger.Warning("New CAS database is empty - keeping old values")
	return casDB{}, casDB{}
}

func isCASBanned(uid int) bool {
	_, found := casdb[uid]
	if found {
		casDatabaseMatch.Inc()
	}
	return found
}

func casWorker() {
	t := time.NewTicker(1 * time.Hour)
	for {
		_, _ = loadCAS()

		// Here we might automatically ban newly CAS-banned users, but for now we limit the bot to
		// react when an user do some action (to avoid flooding Telegram APIs)

		//chats, err := botdb.ListMyChatrooms()
		//if err != nil {
		//	for _, chat := range chats {
		//		settings, err := botdb.GetChatSetting(chat)
		//		if err != nil {
		//			continue
		//		}
		//		if settings.BotEnabled && settings.OnBlacklistCAS.Action != ACTION_NONE {
		//			for uid := range added {
		//				performAction(nil, &tb.User{
		//					ID: uid,
		//				}, settings.OnBlacklistCAS)
		//				time.Sleep(1 * time.Second)
		//			}
		//		}
		//	}
		//}
		<-t.C
	}
}

package main

import (
	"github.com/levigross/grequests"
	"strconv"
	"strings"
	"time"
)

/*
 * This code loads the CAS Database (Combot Anti Spam)
 * You can find more info here: https://combot.org/cas/
 */

type CASDB map[int]int8

var casdb = CASDB{}

// This function downloads and replace current in-memory CAS database
// and returns the list of added and removed items, or nil,nil in case
// of error
func LoadCAS() (CASDB, CASDB) {
	resp, err := grequests.Get("https://combot.org/api/cas/export.csv", &grequests.RequestOptions{
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0",
	})
	if err != nil {
		logger.Critical("CAS database download error: ", err)
		return nil, nil
	}
	if strings.Contains(resp.String(), "cloudflare") {
		logger.Critical("CAS database download error: CloudFlare limited")
		logger.Debug(resp.String())
		return nil, nil
	}

	var newcas = CASDB{}
	var added = CASDB{}
	var deleted = CASDB{}

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
			logger.Critical("Cannot convert ID to Telegram UID: ", err)
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

	casdb = newcas
	logger.Debugf("CAS Database updated - items:%d, added:%d, deleted:%d", len(casdb), len(added), len(deleted))

	return added, deleted
}

func IsCASBanned(uid int) bool {
	_, found := casdb[uid]
	return found
}

func CASWorker() {
	t := time.NewTicker(1 * time.Hour)
	for {
		_, _ = LoadCAS()

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

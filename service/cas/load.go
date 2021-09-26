package cas

import (
	"bufio"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrTimeout           = errors.New("CAS database download error: timeout")
	ErrCloudflareLimited = errors.New("CAS database download error: CloudFlare limited")
)

// Load function downloads and replace current in-memory CAS database
// and returns the list of added and removed items, or nil,nil in case
// of error
func (cas *cas) Load() (DB, DB, error) {
	//startms := time.Now()

	req, err := http.NewRequest(http.MethodGet, "https://api.cas.chat/export.csv", nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")

	resp, err := cas.c.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			cas.logger.WithError(err).Warning("CAS database download timeout")
			// casDatabaseDownloadTime.Set(float64(time.Since(startms) / time.Millisecond))
			return nil, nil, ErrTimeout
		}
		cas.logger.WithError(err).Warning("error during CAS DB download")
		return nil, nil, err
	}

	var newcas = DB{}
	var added = DB{}
	var deleted = DB{}

	// Pre-load deleted array
	for id := range cas.db {
		deleted[id] = 1
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		row := scanner.Text()
		if row == "" {
			continue
		}

		if strings.Contains(row, "cloudflare") {
			cas.logger.WithError(err).Warning("CAS DB download limited by cloudflare")
			return nil, nil, ErrCloudflareLimited
		}

		uid, err := strconv.Atoi(row)
		if err != nil {
			cas.logger.WithError(err).Error("Cannot convert ID to Telegram UID")
		} else {
			// Check if it's new
			_, found := cas.db[uid]
			if !found {
				added[uid] = 1
			}

			newcas[uid] = 1

			// Entry exists in the new version of the DB, so delete it from the "pending delete" list
			delete(deleted, uid)
		}
	}

	//casDatabaseDownloadTime.Set(float64(time.Since(startms) / time.Millisecond))

	if len(newcas) > 0 {
		cas.db = newcas
		//casDatabaseSize.Set(float64(len(cas.db)))
		cas.logger.Debugf("CAS Database updated - items:%d, added:%d, deleted:%d", len(cas.db), len(added), len(deleted))
		return added, deleted, nil
	}
	cas.logger.Warning("New CAS database is empty - keeping old values")
	return DB{}, DB{}, nil
}

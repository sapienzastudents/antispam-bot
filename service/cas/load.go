package cas

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrTimeout           = errors.New("CAS database download error: timeout")
	ErrCloudflareLimited = errors.New("CAS database download error: CloudFlare limited")
)

// Load manually retrieve the datatabase from the combot website and loads it in
// memory, replacing the current database. It can be used when the auto-updater
// is disabled (see New function)
func (cas *cas) Load() error {
	//startms := time.Now()

	// Retrieve the current database in CSV format.
	req, err := http.NewRequest(http.MethodGet, "https://api.cas.chat/export.csv", nil)
	if err != nil {
		return err
	}
	// We need to say to Cloudflare that we're somehow a legit browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")

	resp, err := cas.c.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			cas.logger.WithError(err).Warning("CAS database download timeout")
			// casDatabaseDownloadTime.Set(float64(time.Since(startms) / time.Millisecond))
			return ErrTimeout
		}
		cas.logger.WithError(err).Warning("Failed to download CAS DB")
		return err
	} else if resp.StatusCode != http.StatusOK {
		cas.logger.WithField("http-status", resp.StatusCode).Error("Unexpected HTTP status during CAS DB download")
		return err
	}

	// We calculate the new dictionary as separate entity, because if the CSV
	// file is empty (due to upstream error) we can keep old CAS values
	var newcas = map[int64]int8{}

	// Scan the CSV (which is actually a list of integers, one per line)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		row := scanner.Text()
		// Empty line, skipping
		if row == "" {
			continue
		}

		// Cloudflare CDN is used to distribute the CSV. If Cloudflare somehow
		// "detects" a strange connection, they truncate the file and they put a
		// message there
		if strings.Contains(row, "cloudflare") {
			cas.logger.WithError(err).Warning("CAS DB download limited by cloudflare")
			return ErrCloudflareLimited
		}

		// Try to parse the row as user ID (integer)
		if uid, err := strconv.ParseInt(row, 10, 64); err != nil {
			cas.logger.WithError(err).Error("Failed to convert ID to Telegram UID")
		} else {
			newcas[uid] = 1
		}
	}

	//casDatabaseDownloadTime.Set(float64(time.Since(startms) / time.Millisecond))

	// If we can parse at least one item, use the new database and discard the
	// old one.
	if len(newcas) > 0 {
		cas.db = newcas
		//casDatabaseSize.Set(float64(len(cas.db)))
		cas.logger.WithField("items", len(cas.db)).Debug("CAS Database updated")
		return nil
	}
	// Otherwise, keep the old one
	cas.logger.Warning("New CAS database is empty, keeping old values")
	return nil
}

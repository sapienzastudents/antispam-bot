// Package cas loads the CAS Database (Combot Anti Spam). You can find more info
// here: https://combot.org/cas/
package cas

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// CAS is the object used to query and manage the database.
type CAS interface {
	// Load manually retrieve the datatabase from the combot website and loads
	// it in memory, replacing the current database. It can be used when the
	// auto-updater is disabled (see New function).
	Load() error

	// IsBanned returns true if the given Telegram's ID is present in the DB.
	IsBanned(id int64) bool

	// Close unloads the DB and stops the auto-updater worker, if started.
	Close() error
}

// cas is the concrete type that implements CAS interface.
type cas struct {
	c          *http.Client
	db         map[int64]int8
	logger     logrus.FieldLogger
	workerStop bool // Used to stop the worker (if started).
}

// IsBanned returns true if the given Telegram's ID is present in the DB.
func (cas *cas) IsBanned(uid int64) bool {
	_, found := cas.db[uid]
	return found
}

// Close unloads the DB and stops the auto-updater worker, if started.
func (cas *cas) Close() error {
	cas.workerStop = true
	cas.db = make(map[int64]int8)
	return nil
}

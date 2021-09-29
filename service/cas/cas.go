// Package cas loads the CAS Database (Combot Anti Spam)
// You can find more info here: https://combot.org/cas/
package cas

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

// CAS is the object used to query and manage the database
type CAS interface {
	// Load manually retrieve the datatabase from the combot website and loads it in memory, replacing the current
	// database. It can be used when the auto-updater is disabled (see New function)
	Load() error

	// IsBanned check whether a telegram ID is present in the CAS database
	IsBanned(uid int) bool

	// Close unloads the DB and stops the auto-updater worker, if started
	Close() error
}

type cas struct {
	c          *http.Client
	db         map[int]int8
	logger     logrus.FieldLogger
	workerStop bool
}

// IsBanned check whether a telegram ID is present in the CAS database
func (cas *cas) IsBanned(uid int) bool {
	_, found := cas.db[uid]
	return found
}

// Close unloads the DB and stops the auto-updater worker, if started
func (cas *cas) Close() error {
	cas.workerStop = true
	cas.db = make(map[int]int8)
	return nil
}

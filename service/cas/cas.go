// Package cas loads the CAS Database (Combot Anti Spam)
// You can find more info here: https://combot.org/cas/
package cas

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

type DB map[int]int8

type CAS interface {
	Load() (DB, DB, error)
	IsBanned(uid int) bool
	Close() error
}

type cas struct {
	c          *http.Client
	db         DB
	logger     logrus.FieldLogger
	workerStop bool
}

func (cas *cas) IsBanned(uid int) bool {
	_, found := cas.db[uid]
	/*if found {
		casDatabaseMatch.Inc()
	}*/
	return found
}

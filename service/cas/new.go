package cas

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// New returns a CAS instance.
//
// If autoupdate is true, it launches a goroutine that updates the database in
// background and the given logger is used for debug.
//
// If client is not nil it will be used to override default HTTP client.
func New(autoupdate bool, logger logrus.FieldLogger, client *http.Client) (CAS, error) {
	if client == nil {
		client = &http.Client{Timeout: 1 * time.Minute}
	}
	c := cas{
		c:      client,
		logger: logger,
		db:     make(map[int64]int8),
	}
	if autoupdate {
		go c.worker()
	}
	return &c, nil
}

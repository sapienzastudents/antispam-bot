package cas

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// New builds a new CAS instance. If the autoupdate parameter is true, launches a goroutine for database auto-update in
// background. The logger is used for debug logging. The client parameter can be used to override the default HTTP
// client (if nil, the default http client will be used).
func New(autoupdate bool, logger logrus.FieldLogger, client *http.Client) (CAS, error) {
	if client == nil {
		client = &http.Client{Timeout: 1 * time.Minute}
	}
	c := cas{
		c:      client,
		logger: logger,
		db:     make(map[int]int8),
	}
	if autoupdate {
		go c.worker()
	}
	return &c, nil
}

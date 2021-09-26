package cas

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

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

func (cas *cas) Close() error {
	cas.workerStop = true
	return nil
}

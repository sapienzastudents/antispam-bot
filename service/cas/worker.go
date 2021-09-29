package cas

import "time"

func (cas *cas) worker() {
	t := time.NewTicker(1 * time.Hour)
	for !cas.workerStop {
		_ = cas.Load()
		<-t.C
	}
}

package manager

import (
	"time"

	"github.com/alexanderbez/titan/alerts"
	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/monitor"
)

type Manager struct {
	logger   core.Logger
	monitors []monitor.Monitor
	alerters []alerts.Alerter
	ticker   *time.Ticker
}

func New(logger core.Logger, pollInterval uint, monitors []monitor.Monitor, alerters []alerts.Alerter) Manager {
	return Manager{
		logger:   logger.With("module", "manager"),
		monitors: monitors,
		alerters: alerters,
		ticker:   time.NewTicker(time.Duration(pollInterval) * time.Second),
	}
}

func (mngr Manager) Start() {
	mngr.poll()

	// TODO: Do we need to provide a quit channel?
	for {
		<-mngr.ticker.C
		mngr.poll()
	}
}

func (mngr Manager) poll() {
	mngr.logger.Info("monitoring for new alerts to trigger...")

	for _, mon := range mngr.monitors {
		go func(mon monitor.Monitor) {
			res, err := mon.Exec()
			if err != nil {
				return
			}

			for _, alerter := range mngr.alerters {
				// TODO: Cache/figure out how to not spam
				alerter.Alert(res, mon.Memo())
			}
		}(mon)
	}
}

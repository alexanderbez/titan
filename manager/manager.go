package manager

import (
	"time"

	"github.com/alexanderbez/titan/alerts"
	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/monitor"
)

// Manager implements an monitoring manager that is responsible for executing
// various monitors and alerting a series of alerting targets. For every
// successful monitor whose response/result has not been seen before, will be
// sent to each alert target specified.
type Manager struct {
	db       core.DB
	logger   core.Logger
	monitors []monitor.Monitor
	alerters []alerts.Alerter
	ticker   *time.Ticker
}

func New(
	logger core.Logger, db core.DB, cfg config.Config,
	monitors []monitor.Monitor, alerters []alerts.Alerter,
) Manager {

	return Manager{
		db:       db,
		logger:   logger.With("module", "manager"),
		monitors: monitors,
		alerters: alerters,
		ticker:   time.NewTicker(time.Duration(cfg.PollInterval) * time.Second),
	}
}

// Start is responsible for starting the Manager's poller in a go-routine. It
// will poll every config.PollInterval seconds. Errors are logged but do not
// cause the poller or manager to exit.
func (mngr Manager) Start() {
	mngr.poll()

	for {
		<-mngr.ticker.C
		go mngr.poll()
	}
}

// poll iterates over every monitor and attempts an execution. Upon successful
// execution, the result's ID is checked against the DB. If it has not been
// seen before, it will be sent to each alert target. Any error is logged.
func (mngr Manager) poll() {
	mngr.logger.Info("monitoring for new alerts to trigger...")

	for _, mon := range mngr.monitors {
		res, id, err := mon.Exec()
		if err != nil {
			mngr.logger.Debugf("failed to monitor %s; skipping alert: %v", mon.Name(), err)
			continue
		}

		for _, alerter := range mngr.alerters {
			ok, err := mngr.db.Has(core.BadgerAlertsNamespace, id)
			if !ok && err == nil {
				err := alerter.Alert(res, mon.Memo())
				if err == nil {
					err := mngr.db.Set(core.BadgerAlertsNamespace, id, res)
					if err != nil {
						mngr.logger.Debugf("failed to persist alert: %v", err)
					}
				}
			}
		}
	}
}

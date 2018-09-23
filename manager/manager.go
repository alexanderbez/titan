package manager

import (
	"encoding/json"
	"time"

	"github.com/alexanderbez/titan/alerts"
	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/monitor"
)

var (
	// default alert DB TTL of roughly one month
	alertTTL = 30 * 24 * time.Hour

	// MonitorExecKey defines the database key for persisting the latest monitor
	// execution.
	MonitorExecKey = []byte("latestMonitorExec")
)

type (
	// Manager implements an monitoring manager that is responsible for executing
	// various monitors and alerting a series of alerting targets. For every
	// successful monitor whose response/result has not been seen before, will be
	// sent to each alert target specified.
	Manager struct {
		db       core.DB
		logger   core.Logger
		monitors []monitor.Monitor
		alerters []alerts.Alerter
		ticker   *time.Ticker
	}

	monitorExec struct {
		Timestamp          time.Time `json:"timestamp"`
		FailedMonitors     []string  `json:"failed_monitors"`
		SuccessfulMonitors []string  `json:"successful_monitors"`
		FailedAlerts       []string  `json:"failed_alerts"`
		SuccessfulAlerts   []string  `json:"successful_alerts"`
	}
)

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

func newMonitorExec() *monitorExec {
	return &monitorExec{
		Timestamp:          time.Now().UTC(),
		FailedMonitors:     make([]string, 0),
		SuccessfulMonitors: make([]string, 0),
		FailedAlerts:       make([]string, 0),
		SuccessfulAlerts:   make([]string, 0),
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
	mExec := newMonitorExec()

	for _, mon := range mngr.monitors {
		res, id, err := mon.Exec()
		if err != nil {
			mngr.logger.Debugf("failed to monitor %s; skipping alert: %v", mon.Name(), err)
			mExec.FailedMonitors = append(mExec.FailedMonitors, mon.Name())
		} else {
			// The monitor was successful and but may be regarded as seen before.
			mExec.SuccessfulMonitors = append(mExec.SuccessfulMonitors, mon.Name())

			// Attempt to trigger alert for the monitor's response if it has not been
			// seen before (based on ID).
			for _, alerter := range mngr.alerters {
				ok, err := mngr.db.Has(core.BadgerAlertsNamespace, id)
				if !ok && err == nil {
					// Database successfully checked and no previous monitor response has
					// been found.
					err := alerter.Alert(res, mon.Memo())
					if err != nil {
						mExec.FailedAlerts = append(mExec.FailedAlerts, alerter.Name())
					} else {
						mExec.SuccessfulAlerts = append(mExec.SuccessfulAlerts, alerter.Name())

						// Persist the monitor response by the ID with a TTL to prevent
						// alerting spam.
						err := mngr.db.SetWithTTL(core.BadgerAlertsNamespace, id, res, alertTTL)
						if err != nil {
							mngr.logger.Debugf("failed to persist alert: %v", err)
						}
					}
				}
			}
		}

		err = mngr.saveLatestMonitorExec(mExec)
		if err != nil {
			mngr.logger.Debugf("failed to persist latest monitor execution: %v", err)
		}
	}
}

func (mngr Manager) saveLatestMonitorExec(mExec *monitorExec) error {
	raw, err := json.Marshal(mExec)
	if err != nil {
		return err
	}

	err = mngr.db.Set(core.BadgerMonitorsNamespace, MonitorExecKey, raw)
	if err != nil {
		return err
	}

	return nil
}

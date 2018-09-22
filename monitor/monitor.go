package monitor

import (
	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
)

// Monitor defines an interface that is responsible for monitoring for a
// specific event that will ultimately trigger a potential alert.
type Monitor interface {
	Name() string
	Memo() string
	Exec() (resp, id []byte, err error)
}

// CreateMonitors returns a list of initialized monitors. The exact list of
// created monitors is based upon the enabled monitors in the provided
// configuration which is assumed to have been validated.
func CreateMonitors(cfg config.Config, logger core.Logger) (monitors []Monitor) {
	gpm := NewGovProposalMonitor(
		logger, cfg, GovProposalMonitorName, GovProposalMonitorMemo,
	)

	gvm := NewGovVotingMonitor(
		logger, cfg, GovVotingMonitorName, GovVotingMonitorMemo,
	)

	msm := NewMissingSigMonitor(
		logger, cfg, MissingSigMonitorName, MissingSigMonitorMemo,
	)

	dsm := NewDoubleSignMonitor(
		logger, cfg, DoubleSignMonitorName, DoubleSignMonitorMemo,
	)

	jvm := NewJailedValidatorMonitor(
		logger, cfg, JailedValidatorMonitorName, JailedValidatorMonitorMemo,
	)

	// cfg.Monitors is assumed to have a valid list of enabled monitors
	for _, monitor := range cfg.Monitors {
		switch monitor {
		case config.MonitorAll:
			return []Monitor{gpm, gvm, msm, dsm, jvm}

		case config.MonitorNewProposals:
			monitors = append(monitors, gpm)

		case config.MonitorActiveProposals:
			monitors = append(monitors, gvm)

		case config.MonitorJailedValidators:
			monitors = append(monitors, jvm)

		case config.MonitorDoubleSigning:
			monitors = append(monitors, dsm)

		case config.MonitorMissingSignatures:
			monitors = append(monitors, msm)
		}
	}

	return monitors
}

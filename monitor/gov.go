package monitor

import (
	"crypto/sha256"
	"fmt"

	"github.com/pkg/errors"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var (
	_ Monitor = (*GovProposalMonitor)(nil)
	_ Monitor = (*GovVotingMonitor)(nil)
)

// Governance monitor alert related constants.
const (
	GovProposalMonitorMemo = "New Governance Proposals"
	GovProposalMonitorName = "govProposal/new"
	GovVotingMonitorMemo   = "New Active Governance Proposals"
	GovVotingMonitorName   = "govProposal/voting"
)

const (
	govProposalStatusNew    = "DepositPeriod"
	govProposalStatusVoting = "VotingPeriod"
)

type baseGovMonitor struct {
	codec  *wire.Codec
	logger core.Logger
	cm     *core.ClientManager
	name   string
	memo   string
}

func newBaseGovMonitor(logger core.Logger, cfg config.Config, name, memo string) *baseGovMonitor {
	logger = logger.With("module", name)

	codec := wire.NewCodec()
	gov.RegisterWire(codec)

	return &baseGovMonitor{
		codec:  codec,
		logger: logger,
		cm:     core.NewClientManager(cfg.Network.Clients),
		name:   name,
		memo:   memo,
	}
}

// Name implements the Monitor interface. It returns the monitor's name.
func (gm *baseGovMonitor) Name() string { return gm.name }

// Memo implements the Monitor interface. It returns the monitor's memo.
func (gm *baseGovMonitor) Memo() string { return gm.memo }

func (gm baseGovMonitor) getProposals(url string) (resp []byte, proposals []gov.Proposal, err error) {
	resp, err = core.Request(url, core.RequestGET, nil)
	if err != nil {
		return nil, nil, err
	}

	err = gm.codec.UnmarshalJSON(resp, &proposals)
	if err != nil {
		return nil, nil, err
	}

	return resp, proposals, nil
}

// GovProposalMonitor defines a monitor responsible for monitoring new
// governance proposals.
type GovProposalMonitor struct {
	*baseGovMonitor
}

// NewGovProposalMonitor returns a reference to a new GovProposalMonitor.
func NewGovProposalMonitor(logger core.Logger, cfg config.Config, name, memo string) *GovProposalMonitor {
	return &GovProposalMonitor{newBaseGovMonitor(logger, cfg, name, memo)}
}

// Exec implements the Monitor interface. It will attempt to fetch new
// governance proposals. Upon success, the raw response body and an ID that is
// the SHA256 of the response body will be returned and an error otherwise.
func (gpm *GovProposalMonitor) Exec() (resp, id []byte, err error) {
	url := fmt.Sprintf("%s/gov/proposals?status=%s", gpm.cm.Next(), govProposalStatusNew)
	gpm.logger.Debug("monitoring for new governance proposals")

	resp, proposals, err := gpm.getProposals(url)
	if err != nil {
		gpm.logger.Errorf("failed to monitor for new governance proposals: %v", err)
		return nil, nil, errors.Wrap(err, "failed to monitor for new governance proposals")
	}

	// Do not return a response and ID if no proposals were returned as there is
	// no need to alert.
	if len(proposals) == 0 {
		return nil, nil, errors.New("no proposals returned")
	}

	rawHash := sha256.Sum256(resp)
	id = rawHash[:]

	return resp, id, nil
}

// GovVotingMonitor defines a monitor responsible for monitoring governance
// proposals that are in the voting stage.
type GovVotingMonitor struct {
	*baseGovMonitor
}

// NewGovVotingMonitor returns a reference to a new GovVotingMonitor.
func NewGovVotingMonitor(logger core.Logger, cfg config.Config, name, memo string) *GovVotingMonitor {
	return &GovVotingMonitor{newBaseGovMonitor(logger, cfg, name, memo)}
}

// Exec implements the Monitor interface. It will attempt to fetch governance
// proposals that are in the voting stage. Upon success, the raw response body
// and an ID that is the SHA256 of the response body will be returned and an
// error otherwise.
func (gvm *GovVotingMonitor) Exec() (resp, id []byte, err error) {
	url := fmt.Sprintf("%s/gov/proposals?status=%s", gvm.cm.Next(), govProposalStatusVoting)
	gvm.logger.Debug("monitoring for active governance proposals")

	resp, proposals, err := gvm.getProposals(url)
	if err != nil {
		gvm.logger.Errorf("failed to monitor for active governance proposals: %v", err)
		return nil, nil, errors.Wrap(err, "failed to monitor for active governance proposals")
	}

	// Do not return a response and ID if no proposals were returned as there is
	// no need to alert.
	if len(proposals) == 0 {
		return nil, nil, errors.New("no proposals returned")
	}

	rawHash := sha256.Sum256(resp)
	id = rawHash[:]

	return resp, id, nil

}

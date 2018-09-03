package monitor

import (
	"crypto/sha256"
	"fmt"

	"github.com/alexanderbez/titan/core"
)

var (
	_ Monitor = (*GovProposalMonitor)(nil)
	_ Monitor = (*GovVotingMonitor)(nil)
)

const (
	govProposalStatusNew    = "DepositPeriod"
	govProposalStatusVoting = "VotingPeriod"
)

type baseGovMonitor struct {
	logger core.Logger
	cm     *core.ClientManager
	name   string
	memo   string
}

func newBaseGovMonitor(logger core.Logger, clients []string, name, memo string) *baseGovMonitor {
	return &baseGovMonitor{logger, core.NewClientManager(clients), name, memo}
}

// Name implements the Monitor interface. It returns the monitor's name.
func (gm *baseGovMonitor) Name() string { return gm.name }

// Memo implements the Monitor interface. It returns the monitor's memo.
func (gm *baseGovMonitor) Memo() string { return gm.memo }

// GovProposalMonitor defines a monitor responsible for monitoring new
// governance proposals.
type GovProposalMonitor struct {
	*baseGovMonitor
}

// NewGovProposalMonitor returns a reference to a new GovProposalMonitor.
func NewGovProposalMonitor(logger core.Logger, clients []string, name, memo string) *GovProposalMonitor {
	return &GovProposalMonitor{newBaseGovMonitor(logger, clients, name, memo)}
}

// Exec implements the Monitor interface. It will attempt to fetch new
// governance proposals. Upon success, the raw response body and an ID that is
// the SHA256 of the response body will be returned and an error otherwise.
func (gpm *GovProposalMonitor) Exec() (res, id []byte, err error) {
	url := fmt.Sprintf("%s/gov/proposals?status=%s", gpm.cm.Next(), govProposalStatusNew)
	gpm.logger.Debug(fmt.Sprintf("monitoring for new governance proposals from: %s", url))

	rawBody, err := core.Request(url, core.RequestGET, nil)
	if err != nil {
		gpm.logger.Error(fmt.Sprintf("failed to monitor new governance proposals; error: %v", err))
		return nil, nil, err
	}

	bodyHash := sha256.Sum256(rawBody)
	return rawBody, bodyHash[:], nil
}

// GovVotingMonitor defines a monitor responsible for monitoring governance
// proposals that are in the voting stage.
type GovVotingMonitor struct {
	*baseGovMonitor
}

// NewGovVotingMonitor returns a reference to a new GovVotingMonitor.
func NewGovVotingMonitor(logger core.Logger, clients []string, name, memo string) *GovVotingMonitor {
	return &GovVotingMonitor{newBaseGovMonitor(logger, clients, name, memo)}
}

// Exec implements the Monitor interface. It will attempt to fetch governance
// proposals that are in the voting stage. Upon success, the raw response body
// and an ID that is the SHA256 of the response body will be returned and an
// error otherwise.
func (gpm *GovVotingMonitor) Exec() (res, id []byte, err error) {
	url := fmt.Sprintf("%s/gov/proposals?status=%s", gpm.cm.Next(), govProposalStatusVoting)
	gpm.logger.Debug(fmt.Sprintf("monitoring for governance proposals in voting stage from: %s", url))

	rawBody, err := core.Request(url, core.RequestGET, nil)
	if err != nil {
		gpm.logger.Error(fmt.Sprintf("failed to monitor governance proposals in voting stage; error: %v", err))
		return nil, nil, err
	}

	bodyHash := sha256.Sum256(rawBody)
	return rawBody, bodyHash[:], nil
}

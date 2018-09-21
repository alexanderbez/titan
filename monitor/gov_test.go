package monitor_test

import (
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/monitor"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

func newTestGovProposalMonitor(t *testing.T, ts *httptest.Server) *monitor.GovProposalMonitor {
	logger, err := core.CreateBaseLogger("", false)
	require.NoError(t, err)

	clients := []string{ts.URL}
	cfg := config.Config{Network: config.NetworkConfig{Clients: clients}}

	return monitor.NewGovProposalMonitor(
		logger, cfg, "govProposal/new", "New Governance Proposals",
	)
}

func newTestGovVotingMonitor(t *testing.T, ts *httptest.Server) *monitor.GovVotingMonitor {
	logger, err := core.CreateBaseLogger("", false)
	require.NoError(t, err)

	clients := []string{ts.URL}
	cfg := config.Config{Network: config.NetworkConfig{Clients: clients}}

	return monitor.NewGovVotingMonitor(
		logger, cfg, "govProposal/voting", "New Active Governance Proposals",
	)
}

func TestEmptyProposals(t *testing.T) {
	codec := wire.NewCodec()
	gov.RegisterWire(codec)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proposals []gov.Proposal

		raw, err := codec.MarshalJSON(proposals)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	gpm := newTestGovProposalMonitor(t, ts)

	resp, id, err := gpm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestNewProposals(t *testing.T) {
	codec := wire.NewCodec()
	gov.RegisterWire(codec)

	proposals := []gov.Proposal{
		&gov.TextProposal{
			ProposalID:   1,
			Title:        "test text proposal",
			Description:  "test text proposal",
			ProposalType: gov.ProposalTypeText,
			Status:       gov.StatusDepositPeriod,
			TallyResult:  gov.EmptyTallyResult(),
		},
	}

	raw, err := codec.MarshalJSON(proposals)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	gpm := newTestGovProposalMonitor(t, ts)

	resp, id, err := gpm.Exec()
	require.NoError(t, err)

	var props []gov.Proposal
	err = codec.UnmarshalJSON(resp, &props)
	require.NoError(t, err)

	rawHash := sha256.Sum256(resp)
	exID := rawHash[:]

	require.Equal(t, exID, id)
	require.Len(t, props, len(proposals))
}

func TestEmptyActiveProposals(t *testing.T) {
	codec := wire.NewCodec()
	gov.RegisterWire(codec)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proposals []gov.Proposal

		raw, err := codec.MarshalJSON(proposals)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	gvm := newTestGovVotingMonitor(t, ts)

	resp, id, err := gvm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestNewActiveProposals(t *testing.T) {
	codec := wire.NewCodec()
	gov.RegisterWire(codec)

	proposals := []gov.Proposal{
		&gov.TextProposal{
			ProposalID:   1,
			Title:        "test text proposal",
			Description:  "test text proposal",
			ProposalType: gov.ProposalTypeText,
			Status:       gov.StatusVotingPeriod,
			TallyResult:  gov.EmptyTallyResult(),
		},
	}

	raw, err := codec.MarshalJSON(proposals)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	gvm := newTestGovVotingMonitor(t, ts)

	resp, id, err := gvm.Exec()
	require.NoError(t, err)

	var props []gov.Proposal
	err = codec.UnmarshalJSON(resp, &props)
	require.NoError(t, err)

	rawHash := sha256.Sum256(resp)
	exID := rawHash[:]

	require.Equal(t, exID, id)
	require.Len(t, props, len(proposals))
}

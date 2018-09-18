package monitor_test

import (
	"encoding/hex"
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

	res, id, err := gpm.Exec()
	require.Error(t, err)
	require.Nil(t, res)
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

	res, id, err := gpm.Exec()
	require.NoError(t, err)
	require.Equal(t, res, raw)
	require.Equal(t, hex.EncodeToString(id), "473f109e0729b4751cd59b1350cbe56e931e816d0a907083ddbb7f176f9c1baa")
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

	res, id, err := gvm.Exec()
	require.Error(t, err)
	require.Nil(t, res)
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

	res, id, err := gvm.Exec()
	require.NoError(t, err)
	require.Equal(t, res, raw)
	require.Equal(t, hex.EncodeToString(id), "f9e72639bb3790fff6320897110a47c4e763657a994171adac545c22da18192c")
}

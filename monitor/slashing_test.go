package monitor_test

import (
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/monitor"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func newSlashingTestCodec() *wire.Codec {
	codec := wire.NewCodec()
	stake.RegisterWire(codec)
	ctypes.RegisterAmino(codec)

	return codec
}

func newTestMissingSigMonitor(t *testing.T, cfg config.Config) *monitor.MissingSigMonitor {
	logger, err := core.CreateBaseLogger("", false)
	require.NoError(t, err)

	return monitor.NewMissingSigMonitor(
		logger, cfg, monitor.MissingSigMonitorName, monitor.MissingSigMonitorMemo,
	)
}

func newTestDoubleSignMonitor(t *testing.T, cfg config.Config) *monitor.DoubleSignMonitor {
	logger, err := core.CreateBaseLogger("", false)
	require.NoError(t, err)

	return monitor.NewDoubleSignMonitor(
		logger, cfg, monitor.DoubleSignMonitorName, monitor.DoubleSignMonitorMemo,
	)
}

// 2. Matched missing singers

func TestNoMatchingMissingSignatures(t *testing.T) {
	codec := newSlashingTestCodec()
	pubKey := ed25519.GenPrivKey().PubKey()

	commits := &tmtypes.Commit{
		Precommits: []*tmtypes.Vote{
			&tmtypes.Vote{ValidatorAddress: pubKey.Address()},
		},
	}

	block := &ctypes.ResultBlock{
		Block: tmtypes.MakeBlock(1, nil, commits, []tmtypes.Evidence{}),
	}

	raw, err := codec.MarshalJSON(block)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filter: config.Filter{
			Validators: []config.ValidatorFilter{},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	msm := newTestMissingSigMonitor(t, cfg)

	resp, id, err := msm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestMatchingMissingSignatures(t *testing.T) {
	codec := newSlashingTestCodec()
	pubKey1 := ed25519.GenPrivKey().PubKey()
	pubKey2 := ed25519.GenPrivKey().PubKey()

	commits := &tmtypes.Commit{
		Precommits: []*tmtypes.Vote{
			&tmtypes.Vote{ValidatorAddress: pubKey1.Address()},
		},
	}

	block := &ctypes.ResultBlock{
		Block: tmtypes.MakeBlock(1, nil, commits, []tmtypes.Evidence{}),
	}

	raw, err := codec.MarshalJSON(block)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filter: config.Filter{
			Validators: []config.ValidatorFilter{
				config.ValidatorFilter{Address: pubKey2.Address().String()},
			},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	msm := newTestMissingSigMonitor(t, cfg)

	resp, id, err := msm.Exec()
	require.NoError(t, err)

	var missingSigners monitor.MissingSigners
	err = codec.UnmarshalJSON(resp, &missingSigners)
	require.NoError(t, err)

	rawHash := sha256.Sum256(resp)
	exID := rawHash[:]

	require.Equal(t, exID, id)
	require.Len(t, missingSigners.MissingSigners, 1)
	require.Equal(t, missingSigners.MissingSigners[0], pubKey2.Address().String())
}

func TestNoDoubleSigners(t *testing.T) {
	codec := newSlashingTestCodec()

	block := &ctypes.ResultBlock{
		Block: tmtypes.MakeBlock(1, nil, nil, []tmtypes.Evidence{}),
	}

	raw, err := codec.MarshalJSON(block)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filter: config.Filter{
			Validators: []config.ValidatorFilter{},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	dsm := newTestDoubleSignMonitor(t, cfg)

	resp, id, err := dsm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestNoMatchingDoubleSigners(t *testing.T) {
	codec := newSlashingTestCodec()
	pubKey1 := ed25519.GenPrivKey().PubKey()
	pubKey2 := ed25519.GenPrivKey().PubKey()

	block := &ctypes.ResultBlock{
		Block: tmtypes.MakeBlock(1, nil, nil, []tmtypes.Evidence{
			&tmtypes.DuplicateVoteEvidence{PubKey: pubKey1},
		}),
	}

	raw, err := codec.MarshalJSON(block)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filter: config.Filter{
			Validators: []config.ValidatorFilter{
				config.ValidatorFilter{Address: pubKey2.Address().String()},
			},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	dsm := newTestDoubleSignMonitor(t, cfg)

	resp, id, err := dsm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestMatchingDoubleSigners(t *testing.T) {
	codec := newSlashingTestCodec()
	pubKey1 := ed25519.GenPrivKey().PubKey()

	block := &ctypes.ResultBlock{
		Block: tmtypes.MakeBlock(1, nil, nil, []tmtypes.Evidence{
			&tmtypes.DuplicateVoteEvidence{PubKey: pubKey1},
		}),
	}

	raw, err := codec.MarshalJSON(block)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filter: config.Filter{
			Validators: []config.ValidatorFilter{
				config.ValidatorFilter{Address: pubKey1.Address().String()},
			},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	dsm := newTestDoubleSignMonitor(t, cfg)

	resp, id, err := dsm.Exec()
	require.NoError(t, err)

	var doubleSigners monitor.DoubleSigners
	err = codec.UnmarshalJSON(resp, &doubleSigners)
	require.NoError(t, err)

	rawHash := sha256.Sum256(resp)
	exID := rawHash[:]

	require.Equal(t, exID, id)
	require.Len(t, doubleSigners.DoubleSigners, len(block.Block.Evidence.Evidence))
}

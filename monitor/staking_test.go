package monitor_test

import (
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/alexanderbez/titan/monitor"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"

	"github.com/stretchr/testify/require"
)

func newStakingTestCodec() *wire.Codec {
	codec := wire.NewCodec()
	stake.RegisterWire(codec)
	wire.RegisterCrypto(codec)

	return codec
}

func newTestJailedValidatorMonitor(t *testing.T, cfg config.Config) *monitor.JailedValidatorMonitor {
	logger, err := core.CreateBaseLogger("", false)
	require.NoError(t, err)

	return monitor.NewJailedValidatorMonitor(
		logger, cfg, monitor.JailedValidatorMonitorName, monitor.JailedValidatorMonitorMemo,
	)
}

func TestNoJailedValidators(t *testing.T) {
	codec := newStakingTestCodec()

	opAddr1, err := sdk.AccAddressFromBech32("cosmosaccaddr1chchjxgackcqkn9fqgpsc4n9xamx4flgndapzg")
	require.NoError(t, err)

	validators := []stake.Validator{
		stake.Validator{Owner: opAddr1, Revoked: false},
	}

	raw, err := codec.MarshalJSON(validators)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filters: config.Filters{
			Validators: []config.ValidatorFilter{},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	jvm := newTestJailedValidatorMonitor(t, cfg)

	resp, id, err := jvm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestNoMatchingJailedValidators(t *testing.T) {
	codec := newStakingTestCodec()

	opAddr1, err := sdk.AccAddressFromBech32("cosmosaccaddr1chchjxgackcqkn9fqgpsc4n9xamx4flgndapzg")
	require.NoError(t, err)

	opAddr2, err := sdk.AccAddressFromBech32("cosmosaccaddr1y2z20pwqu5qpclque3pqkguruvheum2djtzjw3")
	require.NoError(t, err)

	validators := []stake.Validator{
		stake.Validator{Owner: opAddr1, Revoked: true},
	}

	raw, err := codec.MarshalJSON(validators)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filters: config.Filters{
			Validators: []config.ValidatorFilter{
				config.ValidatorFilter{Operator: opAddr2.String()},
			},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	jvm := newTestJailedValidatorMonitor(t, cfg)

	resp, id, err := jvm.Exec()
	require.Error(t, err)
	require.Nil(t, resp)
	require.Nil(t, id)
}

func TestMatchingJailedValidators(t *testing.T) {
	codec := newStakingTestCodec()

	opAddr1, err := sdk.AccAddressFromBech32("cosmosaccaddr1chchjxgackcqkn9fqgpsc4n9xamx4flgndapzg")
	require.NoError(t, err)

	val := stake.NewValidator(opAddr1, nil, stake.Description{})
	val.Revoked = true

	validators := []stakeTypes.Validator{val}

	raw, err := wire.MarshalJSONIndent(codec, validators)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(raw)
	}))
	defer ts.Close()

	clients := []string{ts.URL}
	cfg := config.Config{
		Filters: config.Filters{
			Validators: []config.ValidatorFilter{
				config.ValidatorFilter{Operator: opAddr1.String()},
			},
		},
		Network: config.NetworkConfig{Clients: clients},
	}

	jvm := newTestJailedValidatorMonitor(t, cfg)
	resp, id, err := jvm.Exec()
	require.NoError(t, err)

	var vals []stakeTypes.Validator
	err = codec.UnmarshalJSON(resp, &vals)
	require.NoError(t, err)

	rawHash := sha256.Sum256(resp)
	exID := rawHash[:]

	require.Equal(t, exID, id)
	require.Len(t, vals, len(validators))
}

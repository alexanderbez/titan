package monitor

import (
	"crypto/sha256"
	"fmt"

	"github.com/pkg/errors"

	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

var (
	_ Monitor = (*JailedValidatorMonitor)(nil)
)

// Staking monitor alert related constants.
const (
	JailedValidatorMonitorMemo = "New Jailed Validators"
	JailedValidatorMonitorName = "staking/jailed"
)

type baseStakingMonitor struct {
	codec  *wire.Codec
	logger core.Logger
	filter []config.ValidatorFilter
	cm     *core.ClientManager
	name   string
	memo   string
}

func newBaseStakingMonitor(logger core.Logger, cfg config.Config, name, memo string) *baseStakingMonitor {
	logger = logger.With("module", name)

	codec := wire.NewCodec()
	stake.RegisterWire(codec)
	wire.RegisterCrypto(codec)

	return &baseStakingMonitor{
		codec:  codec,
		logger: logger,
		filter: cfg.Filters.Validators,
		cm:     core.NewClientManager(cfg.Network.Clients),
		name:   name,
		memo:   memo,
	}
}

// Name implements the Monitor interface. It returns the monitor's name.
func (sm *baseStakingMonitor) Name() string { return sm.name }

// Memo implements the Monitor interface. It returns the monitor's memo.
func (sm *baseStakingMonitor) Memo() string { return sm.memo }

func (sm baseStakingMonitor) getValidators(url string) (resp []byte, vals []stakeTypes.Validator, err error) {
	resp, err = core.Request(url, core.RequestGET, nil)
	if err != nil {
		return nil, nil, err
	}

	err = sm.codec.UnmarshalJSON(resp, &vals)
	if err != nil {
		return nil, nil, err
	}

	return resp, vals, nil
}

// JailedValidatorMonitor defines a monitor responsible for monitoring when
// certain validators become jailed.
type JailedValidatorMonitor struct {
	*baseStakingMonitor
}

// NewJailedValidatorMonitor returns a reference to a new
// JailedValidatorMonitor.
func NewJailedValidatorMonitor(logger core.Logger, cfg config.Config, name, memo string) *JailedValidatorMonitor {
	return &JailedValidatorMonitor{newBaseStakingMonitor(logger, cfg, name, memo)}
}

// Exec implements the Monitor interface. It attempts to fetch validators that
// are jailed and match against a given filter of validator addresses. Upon
// success, the serialized encoding of the filtered validators and an ID that
// is the SHA256 of said encoding will be returned and an error otherwise.
func (jvm *JailedValidatorMonitor) Exec() (resp, id []byte, err error) {
	url := fmt.Sprintf("%s/stake/validators", jvm.cm.Next())
	jvm.logger.Info("monitoring for new jailed validators")

	_, vals, err := jvm.getValidators(url)
	if err != nil {
		jvm.logger.Errorf("failed to get all validators: %v", err)
		return nil, nil, err
	}

	filtersMap := make(map[string]struct{}, len(jvm.filter))
	for _, validatorFilter := range jvm.filter {
		filtersMap[validatorFilter.Operator] = struct{}{}
	}

	// filter validators that are jailed and match the given filter of addresses
	var filteredVals []stakeTypes.Validator
	for _, val := range vals {
		if _, ok := filtersMap[val.Owner.String()]; ok {
			// TODO: Update once the SDK version has been updated to support the
			// jailed field.
			if val.Revoked {
				filteredVals = append(filteredVals, val)
			}
		}
	}

	// Do not return a response and ID if no validators were returned as there is
	// no need to alert.
	if len(filteredVals) == 0 {
		return nil, nil, errors.New("no validators matching filter returned")
	}

	raw, err := wire.MarshalJSONIndent(jvm.codec, filteredVals)
	if err != nil {
		jvm.logger.Errorf("failed to serialize filtered validators: %v", err)
		return nil, nil, errors.Wrap(err, "failed to serialize filtered validators")
	}

	rawHash := sha256.Sum256(raw)
	id = rawHash[:]

	return raw, id, nil
}

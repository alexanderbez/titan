package monitor

import (
	"crypto/sha256"
	"fmt"

	"github.com/alexanderbez/godash"
	"github.com/alexanderbez/titan/config"
	"github.com/alexanderbez/titan/core"
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	_ Monitor = (*MissingSigMonitor)(nil)
	_ Monitor = (*DoubleSignMonitor)(nil)
)

// Slashing monitor alert related constants.
const (
	MissingSigMonitorMemo = "Missing Signatures From Validators"
	MissingSigMonitorName = "slashing/missingSig"
	DoubleSignMonitorMemo = "Discovered Double Signing Validators"
	DoubleSignMonitorName = "slashing/doubleSign"
)

type (
	baseSlashingMonitor struct {
		codec  *wire.Codec
		logger core.Logger
		filter []config.ValidatorFilter
		cm     *core.ClientManager

		name string
		memo string
	}

	// MissingSigners defines a structure for containing addresses of validators
	// that have missed a signature/precommit for a given block height.
	MissingSigners struct {
		Height         int64    `json:"height"`
		MissingSigners []string `json:"missing_signers"`
	}

	// DoubleSigners defines a structure for containing addresses of validators
	// that have double signed for a given block height.
	DoubleSigners struct {
		Height        int64    `json:"height"`
		DoubleSigners []string `json:"double_signers"`
	}
)

func newBaseSlashingMonitor(logger core.Logger, cfg config.Config, name, memo string) *baseSlashingMonitor {
	logger = logger.With("module", name)

	codec := wire.NewCodec()
	stake.RegisterWire(codec)
	ctypes.RegisterAmino(codec)

	return &baseSlashingMonitor{
		codec:  codec,
		logger: logger,
		filter: cfg.Filters.Validators,
		cm:     core.NewClientManager(cfg.Network.Clients),
		name:   name,
		memo:   memo,
	}
}

// Name implements the Monitor interface. It returns the monitor's name.
func (sm *baseSlashingMonitor) Name() string { return sm.name }

// Memo implements the Monitor interface. It returns the monitor's memo.
func (sm *baseSlashingMonitor) Memo() string { return sm.memo }

func (sm baseSlashingMonitor) getBlock(url string) (resp []byte, block *ctypes.ResultBlock, err error) {
	resp, err = core.Request(url, core.RequestGET, nil)
	if err != nil {
		return nil, nil, err
	}

	err = sm.codec.UnmarshalJSON(resp, &block)
	if err != nil {
		return nil, nil, err
	}

	return resp, block, nil
}

// MissingSigMonitor defines a monitor responsible for monitoring when filtered
// validators fail to sign a block.
type MissingSigMonitor struct {
	*baseSlashingMonitor
}

// NewMissingSigMonitor returns a reference to a new MissingSigMonitor.
func NewMissingSigMonitor(logger core.Logger, cfg config.Config, name, memo string) *MissingSigMonitor {
	return &MissingSigMonitor{newBaseSlashingMonitor(logger, cfg, name, memo)}
}

// Exec implements the Monitor interface. It attempts to fetch validators that
// have missed signing the latest block based on a given filter of validator
// addresses. Any matches are serialized and an ID that is the SHA256 of said
// encoding will be returned and an error otherwise.
func (msm *MissingSigMonitor) Exec() (resp, id []byte, err error) {
	url := fmt.Sprintf("%s/blocks/latest", msm.cm.Next())
	msm.logger.Info("monitoring for validators that have missed signing the latest block")

	_, block, err := msm.getBlock(url)
	if err != nil {
		msm.logger.Errorf("failed to monitor for the latest block: %v", err)
		return nil, nil, errors.Wrap(err, "failed to get latest block")
	}

	filtersMap := make(map[string]struct{}, len(msm.filter))
	for _, validatorFilter := range msm.filter {
		filtersMap[validatorFilter.Address] = struct{}{}
	}

	for _, vote := range block.Block.LastCommit.Precommits {
		if vote != nil {
			delete(filtersMap, vote.ValidatorAddress.String())
		}
	}

	// remaining addresses in filters must have missed signing the latest block
	var missedSigners []string

	if err = godash.MapKeys(filtersMap, &missedSigners); err != nil {
		return nil, nil, errors.Wrap(err, "failed to get missing signers from filter")
	}

	if len(missedSigners) == 0 {
		return nil, nil, errors.New("no validators matching filter returned")
	}

	ms := MissingSigners{
		Height:         block.Block.Header.Height - 1,
		MissingSigners: missedSigners,
	}

	raw, err := wire.MarshalJSONIndent(msm.codec, ms)
	if err != nil {
		msm.logger.Errorf("failed to serialize filtered validators: %v", err)
		return nil, nil, errors.Wrap(err, "failed to serialize filtered validators")
	}

	rawHash := sha256.Sum256(raw)
	id = rawHash[:]

	return raw, id, nil
}

// DoubleSignMonitor defines a monitor responsible for monitoring when filtered
// validators double sign a block.
type DoubleSignMonitor struct {
	*baseSlashingMonitor
}

// NewDoubleSignMonitor returns a reference to a new DoubleSignMonitor.
func NewDoubleSignMonitor(logger core.Logger, cfg config.Config, name, memo string) *DoubleSignMonitor {
	return &DoubleSignMonitor{newBaseSlashingMonitor(logger, cfg, name, memo)}
}

// Exec implements the Monitor interface. It attempts to fetch validators that
// have double signed the latest block and match against a given filter of
// validator addresses. Upon success, the serialized encoding of the filtered
// validator addresses and an ID that is the SHA256 of said encoding will be
// returned and an error otherwise.
func (dsm *DoubleSignMonitor) Exec() (resp, id []byte, err error) {
	url := fmt.Sprintf("%s/blocks/latest", dsm.cm.Next())
	dsm.logger.Info("monitoring for validators that have double signed")

	_, block, err := dsm.getBlock(url)
	if err != nil {
		dsm.logger.Errorf("failed to monitor for the latest block: %v", err)
		return nil, nil, errors.Wrap(err, "failed to get latest block")
	}

	filtersMap := make(map[string]struct{}, len(dsm.filter))
	for _, validatorFilter := range dsm.filter {
		filtersMap[validatorFilter.Address] = struct{}{}
	}

	var byzantineAddrs []string

	for _, e := range block.Block.Evidence.Evidence {
		dve, ok := e.(*tmtypes.DuplicateVoteEvidence)
		if ok && dve != nil {
			valAddr := dve.PubKey.Address().String()

			// check the byzantine signer against the filter map of addresses
			if _, ok := filtersMap[valAddr]; ok {
				byzantineAddrs = append(byzantineAddrs, valAddr)
			}
		}
	}

	if len(byzantineAddrs) == 0 {
		return nil, nil, errors.New("no validators matching filter returned")
	}

	ds := DoubleSigners{
		Height:        block.Block.Header.Height - 1,
		DoubleSigners: byzantineAddrs,
	}

	raw, err := wire.MarshalJSONIndent(dsm.codec, ds)
	if err != nil {
		dsm.logger.Errorf("failed to serialize filtered validators: %v", err)
		return nil, nil, errors.Wrap(err, "failed to serialize filtered validators")
	}

	rawHash := sha256.Sum256(raw)
	id = rawHash[:]

	return raw, id, nil
}

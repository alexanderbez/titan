package monitor

import (
	"crypto/sha256"
	"fmt"

	"github.com/alexanderbez/godash"
	"github.com/alexanderbez/titan/core"
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var (
	_ Monitor = (*MissedBlocksMonitor)(nil)
)

type (
	baseSlashingMonitor struct {
		codec  *wire.Codec
		logger core.Logger
		cm     *core.ClientManager
		name   string
		memo   string
	}

	missingSigners struct {
		Height         int64    `json:"height"`
		MissingSigners []string `json:"missing_signers"`
	}
)

func newBaseSlashingMonitor(logger core.Logger, clients []string, name, memo string) *baseSlashingMonitor {
	logger = logger.With("module", name)

	codec := wire.NewCodec()
	stake.RegisterWire(codec)
	ctypes.RegisterAmino(codec)

	return &baseSlashingMonitor{
		codec:  codec,
		logger: logger,
		cm:     core.NewClientManager(clients),
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

// MissedBlocksMonitor defines a monitor responsible for monitoring when
// certain validators miss a block.
type MissedBlocksMonitor struct {
	*baseSlashingMonitor
}

// NewMissedBlocksMonitor returns a reference to a new MissedBlocksMonitor.
func NewMissedBlocksMonitor(logger core.Logger, clients []string, name, memo string) *MissedBlocksMonitor {
	return &MissedBlocksMonitor{newBaseSlashingMonitor(logger, clients, name, memo)}
}

// Exec implements the Monitor interface. It attempts to fetch validators that
// have missed signing the latest block based on a given filter of validator
// addresses. Any matches are serialized and an ID that is the SHA256 of said
// encoding will be returned and an error otherwise.
func (mbm *MissedBlocksMonitor) Exec(filters []string) (resp, id []byte, err error) {
	url := fmt.Sprintf("%s/blocks/312114", mbm.cm.Next())
	mbm.logger.Debug("monitoring for validators that have missed the latest block")

	_, block, err := mbm.getBlock(url)
	if err != nil {
		mbm.logger.Errorf("failed to monitor for the latest block: %v", err)
		return nil, nil, errors.Wrap(err, "failed to get latest block")
	}

	filtersMap := make(map[string]struct{}, len(filters))
	for _, filter := range filters {
		if filter != "" {
			filtersMap[filter] = struct{}{}
		}
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

	ms := missingSigners{
		Height:         block.BlockMeta.Header.Height - 1,
		MissingSigners: missedSigners,
	}

	raw, err := wire.MarshalJSONIndent(mbm.codec, ms)
	if err != nil {
		mbm.logger.Errorf("failed to serialize filtered validators: %v", err)
		return nil, nil, errors.Wrap(err, "failed to serialize filtered validators")
	}

	rawHash := sha256.Sum256(raw)
	id = rawHash[:]

	return raw, id, nil
}

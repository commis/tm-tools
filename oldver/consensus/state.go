package consensus

import (
	"fmt"
	"hash/crc32"
	"time"

	cstypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/p2p"
)

var crc32c = crc32.MakeTable(crc32.Castagnoli)

// msgs from the reactor which may update the state
type msgInfo struct {
	Msg    ConsensusMessage `json:"msg"`
	PeerID p2p.ID           `json:"peer_key"`
}

// internally generated messages which may update the state
type timeoutInfo struct {
	Duration time.Duration         `json:"duration"`
	Height   int64                 `json:"height"`
	Round    int                   `json:"round"`
	Step     cstypes.RoundStepType `json:"step"`
}

func (ti *timeoutInfo) String() string {
	return fmt.Sprintf("%v ; %d/%d %v", ti.Duration, ti.Height, ti.Round, ti.Step)
}

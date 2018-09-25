package convert

import (
	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/types"
)

func NewHeader(o *his.Header, lastBlockId *types.BlockID) types.Header {
	n := types.Header{}
	n.ChainID = o.ChainID
	n.Height = o.Height
	n.Time = o.Time
	n.NumTxs = o.NumTxs
	if lastBlockId != nil {
		n.LastBlockID = *lastBlockId
	} else {
		n.LastBlockID = NewBlockID(&o.LastBlockID)
	}
	n.TotalTxs = o.TotalTxs
	n.LastCommitHash = o.LastCommitHash.Bytes()
	n.DataHash = o.DataHash.Bytes()
	n.ValidatorsHash = o.ValidatorsHash.Bytes()
	n.ConsensusHash = o.ConsensusHash.Bytes()
	n.AppHash = o.AppHash.Bytes()
	n.LastResultsHash = o.LastResultsHash.Bytes()
	n.EvidenceHash = o.EvidenceHash.Bytes()

	return n
}

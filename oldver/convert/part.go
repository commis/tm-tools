package convert

import (
	"github.com/commis/tm-tools/oldver"
	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

func NewBlockFromOld(ldb dbm.DB, height int64, lastBlockID *types.BlockID, nState *state.State) *types.Block {
	oBlock := oldver.LoadOldBlock(ldb, height)
	if oBlock != nil {
		nBlock := &types.Block{}
		nBlock.Data = NewData(oBlock.Data)
		nBlock.LastCommit = NewCommit(oBlock.LastCommit, lastBlockID, nState)
		nBlock.Header = NewHeader(oBlock.Header, lastBlockID)
		return nBlock
	}
	return nil
}

func NewData(o *his.Data) types.Data {
	txs := []types.Tx{}
	for _, tx := range o.Txs {
		txs = append(txs, []byte(tx))
	}
	return types.Data{
		Txs: txs,
	}
}

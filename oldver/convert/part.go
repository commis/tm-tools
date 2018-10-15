package convert

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tendermint/tendermint/privval"

	his "github.com/commis/tm-tools/oldver/types"
	oldtype "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func LoadOldBlock(ldb dbm.DB, height int64) *oldtype.Block {
	meta := LoadOldBlockMeta(ldb, height)

	bytez := []byte{}
	for i := 0; i < meta.BlockID.PartsHeader.Total; i++ {
		part := LoadOldBlockPart(ldb, height, i)
		bytez = append(bytez, part.Bytes...)
	}

	var n int
	var err error
	block := wire.ReadBinary(&oldtype.Block{}, bytes.NewReader(bytez), 0, &n, &err).(*oldtype.Block)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error reading block: %v", err))
	}
	return block
}

func LoadOldBlockMeta(ldb dbm.DB, height int64) *oldtype.BlockMeta {
	var n int
	var err error
	r := GetReader(ldb, calcBlockMetaKey(height))
	meta := wire.ReadBinary(&oldtype.BlockMeta{}, r, 0, &n, &err).(*oldtype.BlockMeta)
	if err != nil {
		cmn.Exit(fmt.Sprintf(err.Error()))
	}
	return meta
}

func LoadOldBlockPart(ldb dbm.DB, height int64, index int) *oldtype.Part {
	buf := ldb.Get(calcBlockPartKey(height, index))
	if buf == nil {
		return nil
	}

	var n int
	var err error
	part := wire.ReadBinary(&oldtype.Part{}, bytes.NewReader(buf), 0, &n, &err).(*oldtype.Part)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error reading block part: %v", err))
	}
	return part
}

func NewBlockFromOld(ldb dbm.DB, height int64, lastBlockID *types.BlockID, nState *state.State, pv *privval.FilePV) *types.Block {
	oBlock := LoadOldBlock(ldb, height)
	if oBlock == nil {
		return nil
	}

	nBlock := &types.Block{}
	nBlock.Header = NewHeader(oBlock.Header, lastBlockID)
	nBlock.Data = NewData(oBlock.Data)

	nBlock.LastCommit = NewCommit(oBlock.LastCommit, nState, pv)

	return nBlock
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

func GetReader(ldb dbm.DB, key []byte) io.Reader {
	bytez := ldb.Get(key)
	if bytez == nil {
		return nil
	}
	return bytes.NewReader(bytez)
}

//==============================================================================
func calcBlockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

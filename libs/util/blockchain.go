package util

import (
	"fmt"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/blockchain"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

var (
	BlockStoreKey    = "blockStore"
	StateKey         = "stateKey"
	ABCIResponsesKey = "abciResponsesKey"
)

var cdc = amino.NewCodec()

func init() {
	blockchain.RegisterBlockchainMessages(cdc)
	types.RegisterBlockAmino(cdc)
}

// blockstore
func SaveNewBlockStoreStateJSON2(batch db.Batch, totalHeight int64) {
	bsj := blockchain.BlockStoreStateJSON{
		Height: totalHeight,
	}
	bytes, err := cdc.MarshalJSON(bsj)
	if err != nil {
		cmn.PanicSanity(cmn.Fmt("Could not marshal state bytes: %v", err))
	}
	batch.Set([]byte(BlockStoreKey), bytes)
}

// state
func LoadNewState(ldb db.DB) *state.State {
	buf := ldb.Get([]byte(StateKey))
	if len(buf) == 0 {
		return nil
	}

	s := &state.State{}
	err := cdc.UnmarshalBinaryBare(buf, s)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		cmn.Exit(cmn.Fmt(`LoadState: Data has been corrupted or its spec has changed:%v\n`, err))
	}
	return s
}

func SaveNewState(ldb db.DB, s *state.State) {
	buf := s.Bytes()
	ldb.Set([]byte(StateKey), buf)
}

// block header
func LoadNewBlockMeta(ldb db.DB, height int64) *types.BlockMeta {
	buf := ldb.Get(calcBlockMetaKey(height))
	if len(buf) == 0 {
		return nil
	}

	var blockMeta = new(types.BlockMeta)
	err := cdc.UnmarshalBinaryBare(buf, blockMeta)
	if err != nil {
		panic(cmn.ErrorWrap(err, "Error reading block meta"))
	}
	return blockMeta
}

func SaveNewBlockMeta2(batch db.Batch, height int64, blockMeta *types.BlockMeta) {
	metaBytes := cdc.MustMarshalBinaryBare(blockMeta)
	batch.Set(calcBlockMetaKey(height), metaBytes)
}

// block
func LoadNewBlock(ldb db.DB, height int64) *types.Block {
	meta := LoadNewBlockMeta(ldb, height)
	if meta == nil {
		return nil
	}

	buf := []byte{}
	for i := 0; i < meta.BlockID.PartsHeader.Total; i++ {
		part := LoadNewBlockPart(ldb, height, i)
		buf = append(buf, part.Bytes...)
	}

	block := &types.Block{}
	err := cdc.UnmarshalBinary(buf, block)
	if err != nil {
		// NOTE: The existence of meta should imply the existence of the
		// block. So, make sure meta is only saved after blocks are saved.
		panic(cmn.ErrorWrap(err, "Error reading block"))
	}
	return block
}

// block part
func LoadNewBlockPart(ldb db.DB, height int64, index int) *types.Part {
	buf := ldb.Get(calcBlockPartKey(height, index))
	if len(buf) == 0 {
		return nil
	}
	var part = new(types.Part)
	err := cdc.UnmarshalBinaryBare(buf, part)
	if err != nil {
		panic(cmn.ErrorWrap(err, "Error reading block part"))
	}
	return part
}

func SaveNewBlockParts2(batch db.Batch, height int64, block *types.Block, state *state.State) {
	blockParts := block.MakePartSet(state.ConsensusParams.BlockPartSizeBytes)
	for index := 0; index < blockParts.Total(); index++ {
		SaveNewBlockPart2(batch, height, index, blockParts.GetPart(index))
	}
}

func SaveNewBlockPart2(batch db.Batch, height int64, index int, part *types.Part) {
	partBytes := cdc.MustMarshalBinaryBare(part)
	batch.Set(calcBlockPartKey(height, index), partBytes)
}

// block commit
func LoadNewBlockCommit(ldb db.DB, height int64, prefix string) *types.Commit {
	var buf []byte

	if prefix == "C" {
		buf = ldb.Get(calcBlockCommitKey(height))
	} else if prefix == "SC" {
		buf = ldb.Get(calcSeenCommitKey(height))
	}

	blockCommit := &types.Commit{}
	err := cdc.UnmarshalBinaryBare(buf, blockCommit)
	if err != nil {
		panic(cmn.ErrorWrap(err, "Error reading block commit"))
	}
	return blockCommit
}

func SaveNewCommit2(batch db.Batch, height int64, prefix string, commit *types.Commit) {
	var key []byte
	switch prefix {
	case "C":
		key = calcBlockCommitKey(height)
	case "SC":
		key = calcSeenCommitKey(height)
	default:
		panic(prefix)
	}

	buf := cdc.MustMarshalBinaryBare(commit)
	batch.Set(key, buf)
}

// ABCIResponse
func LoadNewABCIResps(ldb db.DB) *state.ABCIResponses {
	buf := ldb.Get([]byte(ABCIResponsesKey))
	resps := &state.ABCIResponses{}
	err := cdc.UnmarshalBinaryBare(buf, resps)
	if err != nil {
		cmn.Exit(cmn.Fmt(`LoadABCIResponses: Data has been corrupted or its spec has changed: %v\n`, err))
	}
	return resps
}

// Evidence
func SaveNewEvidence(batch db.Batch, key []byte, ei interface{}) {
	eiBytes := cdc.MustMarshalBinaryBare(ei)
	batch.Set(key, eiBytes)
}

//==============================================================================
func calcBlockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

func calcBlockCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

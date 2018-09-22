package util

import (
	"fmt"

	his "github.com/commis/tm-tools/oldver/types"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/blockchain"
	"github.com/tendermint/tendermint/evidence"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	BlockStoreKey      = "blockStore"
	StateKey           = "stateKey"
	GenesisDoc         = "genesisDoc"
	ABCIResponsesKey   = "abciResponsesKey"
	ConsensusParamsKey = "consensusParamsKey"
	ValidatorsKey      = "validatorsKey"
)

var cdc = amino.NewCodec()

func init() {
	blockchain.RegisterBlockchainMessages(cdc)
	types.RegisterBlockAmino(cdc)
}

func LoadNewGenesisDoc(db db.DB) *types.GenesisDoc {
	bytes := db.Get([]byte(GenesisDoc))
	if len(bytes) == 0 {
		return nil
	}
	var genDoc *types.GenesisDoc
	err := cdc.UnmarshalJSON(bytes, &genDoc)
	if err != nil {
		fmt.Printf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, bytes)
		return nil
	}
	return genDoc
}

// blockstore
func LoadNewTotalHeight(ldb db.DB) int64 {
	blockStore := LoadNewBlockStoreStateJSON(ldb)
	return blockStore.Height
}

func LoadNewBlockStoreStateJSON(ldb db.DB) blockchain.BlockStoreStateJSON {
	bytes := ldb.Get([]byte(BlockStoreKey))
	if bytes == nil {
		return blockchain.BlockStoreStateJSON{
			Height: 0,
		}
	}
	bsj := blockchain.BlockStoreStateJSON{}
	err := cdc.UnmarshalJSON(bytes, &bsj)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Could not unmarshal bytes: %X", bytes))
	}
	return bsj
}

func SaveNewBlockStoreStateJSON(ldb db.DB, totalHeight int64) {
	bsj := blockchain.BlockStoreStateJSON{
		Height: totalHeight,
	}
	bytes, err := cdc.MarshalJSON(bsj)
	if err != nil {
		cmn.PanicSanity(cmn.Fmt("Could not marshal state bytes: %v", err))
	}
	ldb.SetSync([]byte(BlockStoreKey), bytes)
}

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

func SaveOldState(ldb dbm.DB, s *his.State) {
	buf := s.Bytes()
	ldb.Set([]byte(StateKey), buf)
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
		cmn.Exit(fmt.Sprintf("Error reading block meta, reason: %s", err.Error()))
	}
	return blockMeta
}

func SaveNewBlockMeta2(batch db.Batch, height int64, blockMeta *types.BlockMeta) {
	metaBytes := cdc.MustMarshalBinaryBare(blockMeta)
	batch.Set(calcBlockMetaKey(height), metaBytes)
}

func DeleteBlockMeta(newVer bool, ldb dbm.DB, ndb db.DB, height int64) {
	key := calcBlockMetaKey(height)
	if newVer {
		if ndb.Has(key) {
			ndb.DeleteSync(key)
		}
	} else {
		if ldb.Has(key) {
			ldb.DeleteSync(key)
		}
	}
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
		cmn.Exit(fmt.Sprintf("Error reading block, reason: %s", err.Error()))
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
		cmn.Exit(fmt.Sprintf("Error reading block part, reason: %s", err.Error()))
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

func DeleteOldBlockParts(ldb dbm.DB, block *types.Block, state *his.State) {
	blockParts := block.MakePartSet(state.ConsensusParams.BlockPartSizeBytes)
	for index := 0; index < blockParts.Total(); index++ {
		key := calcBlockPartKey(block.Height, index)
		ldb.Delete(key)
	}
}

func DeleteNewBlockParts(ldb db.DB, block *types.Block, state *state.State) {
	blockParts := block.MakePartSet(state.ConsensusParams.BlockPartSizeBytes)
	for index := 0; index < blockParts.Total(); index++ {
		key := calcBlockPartKey(block.Height, index)
		ldb.Delete(key)
	}
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
		cmn.Exit(fmt.Sprintf("Error reading block commit, reason: %s", err.Error()))
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
		cmn.Exit(prefix)
	}

	buf := cdc.MustMarshalBinaryBare(commit)
	batch.Set(key, buf)
}

func DeleteCommit(newVer bool, ldb dbm.DB, ndb db.DB, height int64) {
	if newVer {
		if key := calcBlockCommitKey(height); ndb.Has(key) {
			ndb.DeleteSync(key)
		}

		if key := calcSeenCommitKey(height); ndb.Has(key) {
			ndb.DeleteSync(key)
		}
	} else {
		if key := calcBlockCommitKey(height); ldb.Has(key) {
			ldb.DeleteSync(key)
		}

		if key := calcSeenCommitKey(height); ldb.Has(key) {
			ldb.DeleteSync(key)
		}
	}
}

// ABCIResponse
func LoadNewABCIResponse(ldb db.DB, height int64) *state.ABCIResponses {
	buf := ldb.Get(CalcABCIResponsesKey(height))
	resps := &state.ABCIResponses{}
	err := cdc.UnmarshalBinaryBare(buf, resps)
	if err != nil {
		//fmt.Printf("LoadABCIResponses: Data has been corrupted or its spec has changed: %v\n", err.Error())
		return nil
	}
	return resps
}

func LoadNewConsensusParamsInfo(ldb db.DB, height int64) *state.ConsensusParamsInfo {
	buf := ldb.Get(CalcConsensusParamsKey(height))
	if len(buf) == 0 {
		return nil
	}

	paramsInfo := new(state.ConsensusParamsInfo)
	err := cdc.UnmarshalBinaryBare(buf, paramsInfo)
	if err != nil {
		//fmt.Sprintf("LoadConsensusParams: Data has been corrupted or its spec has changed: %v\n", err.Error())
		return nil
	}

	return paramsInfo
}

func LoadNewValidatorsInfo(ldb db.DB, height int64) *state.ValidatorsInfo {
	buf := ldb.Get(CalcValidatorsKey(height))
	if len(buf) == 0 {
		return nil
	}

	v := new(state.ValidatorsInfo)
	err := cdc.UnmarshalBinaryBare(buf, v)
	if err != nil {
		cmn.Exit(fmt.Sprintf("LoadValidators: Data has been corrupted or its spec has changed: %v\n", err.Error()))
	}

	return v
}

// Evidence
func SaveNewEvidence(batch db.Batch, key []byte, ei *evidence.EvidenceInfo) {
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

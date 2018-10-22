package op

import (
	"bytes"
	"fmt"

	"github.com/commis/tm-tools/libs/log"
	"github.com/commis/tm-tools/libs/op/hold"
	cvt "github.com/commis/tm-tools/oldver/convert"
	otp "github.com/commis/tm-tools/oldver/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

type TmDataStore struct {
	oldPath     string
	newPath     string
	totalHeight int64
	newState    *state.State

	oBlockDb, oStateDb dbm.DB
	nBlockDb, nStateDb db.DB
}

func CreateTmDataStore(oldPath, newPath string) *TmDataStore {
	ts := &TmDataStore{oldPath: oldPath, newPath: newPath}

	if oldPath != "" {
		dataPath := oldPath + "/data"
		ts.oBlockDb = dbm.NewDB("blockstore", dbm.LevelDBBackend, dataPath)
		ts.oStateDb = dbm.NewDB("state", dbm.LevelDBBackend, dataPath)
	}

	if newPath != "" {
		dataPath := newPath + "/data"
		ts.nBlockDb = db.NewDB("blockstore", db.LevelDBBackend, dataPath)
		ts.nStateDb = db.NewDB("state", db.LevelDBBackend, dataPath)
	}

	return ts
}

func (ts *TmDataStore) OnStop() {
	CloseDbm(ts.oBlockDb)
	CloseDbm(ts.oStateDb)

	CloseDb(ts.nBlockDb)
	CloseDb(ts.nStateDb)
}

func (ts *TmDataStore) GetTotalHeight(isNew bool) {
	if isNew {
		ts.totalHeight = hold.LoadNewBlockHeight(ts.nBlockDb)
	} else {
		ts.totalHeight = cvt.LoadOldTotalHeight(ts.oBlockDb)
	}
	log.Infof("total height: %d", ts.totalHeight)
}

func (ts *TmDataStore) CheckNeedUpgrade(privVal *privval.FilePV) bool {
	nPvFile := ts.newPath + "/config/priv_validator.json"
	nPrivVal := privval.LoadFilePV(nPvFile)
	if bytes.Equal(privVal.Address, nPrivVal.Address) {
		return true
	}

	return false
}

func (ts *TmDataStore) OnBlockStore(startHeight int64) {
	if startHeight < 1 {
		cmn.Exit(fmt.Sprintf("Invalid start height"))
	}

	ts.newState = cvt.InitState(ts.oStateDb)
	lastBlockID := ts.getUpgradeLastBlockID(startHeight)

	cnt := 0
	limit := 1000
	index := startHeight
	batch := ts.nBlockDb.NewBatch()
	for ; index <= ts.totalHeight; index++ {
		cnt++
		ts.newState.LastBlockID = *lastBlockID
		log.Debugf("upgrade block %d", index)
		nBlock := cvt.NewBlockFromOld(ts.oBlockDb, index, lastBlockID)

		blockParts := nBlock.MakePartSet(ts.newState.ConsensusParams.BlockPartSizeBytes)
		nMeta := types.NewBlockMeta(nBlock, blockParts)

		seenCommit := cvt.NewSeenCommit(ts.oBlockDb, nBlock)
		ts.saveBlock(batch, nBlock, nMeta, seenCommit)
		if cnt%limit == 0 {
			log.Infof("batch write %v/%v", cnt, ts.totalHeight)
			batch.WriteSync()
			batch = ts.nBlockDb.NewBatch()
		}

		lastBlockID = &nMeta.BlockID
		ts.upgradeStateData(index)
	}
	if cnt%limit != 0 {
		log.Infof("batch write %v/%v", cnt, ts.totalHeight)
		batch.WriteSync()
	}
	ts.upgradeStateData(index)
	hold.SaveNewState(ts.nStateDb, ts.newState)
}

func (ts *TmDataStore) getUpgradeLastBlockID(sHeight int64) *types.BlockID {
	var lastBlockID *types.BlockID
	if sHeight == 1 {
		lastBlockID = &types.BlockID{}
	} else {
		nMeta := hold.LoadNewBlockMeta(ts.nBlockDb, sHeight-1)
		lastBlockID = &nMeta.BlockID
	}

	return lastBlockID
}

func (ts *TmDataStore) UpdateGenesisDocInStateDB() {
	nGenFile := ts.newPath + "/config/genesis.json"
	nGenDoc, err := types.GenesisDocFromFile(nGenFile)
	if err == nil {
		hold.SaveNewGenesisDoc(ts.nStateDb, nGenDoc)
	}
}

func (ts *TmDataStore) upgradeStateData(height int64) {
	cvt.SaveNewABCIResponse(ts.oStateDb, ts.nStateDb, height)
	cvt.SaveNewConsensusParams(ts.oStateDb, ts.nStateDb, height)
	cvt.SaveNewValidators(ts.oStateDb, ts.nStateDb, height)
}

func (ts *TmDataStore) saveBlock(batch dbm.Batch, block *types.Block, blockMeta *types.BlockMeta, seenCommit *types.Commit) {
	height := block.Height
	hold.SaveNewBlockMeta2(batch, height, blockMeta)
	hold.SaveNewBlockParts2(batch, height, block, ts.newState)
	hold.SaveNewCommit2(batch, height-1, "C", block.LastCommit)
	hold.SaveNewCommit2(batch, height, "SC", seenCommit)
	hold.SaveNewBlockStoreStateJSON2(batch, height)
}

func (ts *TmDataStore) OnBlockRecover(newVersion bool, resetHeight int64) {
	if ts.totalHeight <= resetHeight {
		cmn.Exit(fmt.Sprintf("reset height %d >= total height %d", resetHeight, ts.totalHeight))
	}

	if !newVersion {
		ts.resetOldBlock(resetHeight)
	} else {
		ts.resetNewBlock(resetHeight)
	}
}

func (ts *TmDataStore) resetOldBlock(height int64) {
	state := ts.getOldState(height)
	hold.SaveOldState(ts.oStateDb, state)

	for i := height + 1; i <= ts.totalHeight; i++ {
		block := cvt.LoadOldBlock(ts.oBlockDb, i)
		ts.deleteOldBlock(block, state)
	}

	json := otp.BlockStoreStateJSON{Height: height}
	cvt.SaveOldBlockStoreStateJson(ts.oBlockDb, json)
}

func (ts *TmDataStore) resetNewBlock(height int64) {
	state := ts.getNewState(height)
	hold.SaveNewState(ts.nStateDb, state)

	for i := height + 1; i <= ts.totalHeight; i++ {
		block := hold.LoadNewBlock(ts.nBlockDb, i)
		ts.deleteNewBlock(block, state)
	}

	hold.SaveNewBlockStoreStateJSON(ts.nBlockDb, height)
}

func (ts *TmDataStore) getOldState(height int64) *otp.State {
	state := new(otp.State)

	oState := cvt.LoadOldState(ts.oStateDb)
	lastBlock := cvt.LoadOldBlock(ts.oBlockDb, height-1)
	block := cvt.LoadOldBlock(ts.oBlockDb, height)

	state.ChainID = block.ChainID
	state.LastBlockHeight = lastBlock.Height
	state.LastBlockTotalTx = lastBlock.TotalTxs
	state.LastBlockID = block.LastBlockID
	state.LastBlockTime = lastBlock.Time

	state.Validators = oState.Validators.Copy()
	state.LastValidators = oState.Validators.Copy()
	state.LastHeightValidatorsChanged = oState.LastHeightValidatorsChanged

	state.ConsensusParams = oState.ConsensusParams
	state.LastHeightConsensusParamsChanged = oState.LastHeightConsensusParamsChanged

	state.LastResultsHash = block.LastResultsHash
	state.AppHash = block.AppHash

	return state
}

func (ts *TmDataStore) getNewState(height int64) *state.State {
	state := new(state.State)

	oState := hold.LoadNewState(ts.nStateDb)
	lastBlock := hold.LoadNewBlock(ts.nBlockDb, height-1)
	block := hold.LoadNewBlock(ts.nBlockDb, height)

	state.ChainID = block.ChainID
	state.LastBlockHeight = lastBlock.Height
	state.LastBlockTotalTx = lastBlock.TotalTxs
	state.LastBlockID = block.LastBlockID
	state.LastBlockTime = lastBlock.Time

	state.Validators = oState.Validators.Copy()
	state.LastValidators = oState.Validators.Copy()
	state.LastHeightValidatorsChanged = oState.LastHeightValidatorsChanged

	state.ConsensusParams = oState.ConsensusParams
	state.LastHeightConsensusParamsChanged = oState.LastHeightConsensusParamsChanged

	state.LastResultsHash = block.LastResultsHash
	state.AppHash = block.AppHash

	return state
}

func (ts *TmDataStore) deleteOldBlock(block *otp.Block, state *otp.State) {
	// block
	hold.DeleteBlockMeta(false, ts.oBlockDb, ts.nBlockDb, block.Height)
	hold.DeleteOldBlockParts(ts.oBlockDb, block, state)
	hold.DeleteCommit(false, ts.oBlockDb, ts.nBlockDb, block.Height-1)
	hold.DeleteCommit(false, ts.oBlockDb, ts.nBlockDb, block.Height)

	// state
	stateHeight := block.Height + 1
	cvt.DeleteABCIResponse(false, ts.oStateDb, ts.nStateDb, block.Height)
	cvt.DeleteConsensusParam(false, ts.oStateDb, ts.nStateDb, stateHeight)
	cvt.DeleteValidator(false, ts.oStateDb, ts.nStateDb, stateHeight)
}

func (ts *TmDataStore) deleteNewBlock(block *types.Block, state *state.State) {
	// block
	hold.DeleteBlockMeta(true, ts.oBlockDb, ts.nBlockDb, block.Height)
	hold.DeleteNewBlockParts(ts.nBlockDb, block, state)
	hold.DeleteCommit(true, ts.oBlockDb, ts.nBlockDb, block.Height-1)
	hold.DeleteCommit(true, ts.oBlockDb, ts.nBlockDb, block.Height)

	// state
	stateHeight := block.Height + 1
	cvt.DeleteABCIResponse(true, ts.oStateDb, ts.nStateDb, block.Height)
	cvt.DeleteConsensusParam(true, ts.oStateDb, ts.nStateDb, stateHeight)
	cvt.DeleteValidator(true, ts.oStateDb, ts.nStateDb, stateHeight)
}

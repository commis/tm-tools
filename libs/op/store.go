package op

import (
	"fmt"
	"log"

	"github.com/commis/tm-tools/libs/util"
	cvt "github.com/commis/tm-tools/oldver/convert"
	his "github.com/commis/tm-tools/oldver/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

type TmDataStore struct {
	ethDbPath                                 string
	walPath                                   string
	cfgPath                                   string
	oBlockDb, oStateDb, oEvidenceDb, oTrustDb dbm.DB
	nBlockDb, nStateDb, nEvidenceDb, nTrustDb db.DB
	newState                                  *state.State
	totalHeight                               int64
}

func (c *TmDataStore) OnStart(oldPath, newPath string) {
	if oldPath != "" {
		c.ethDbPath = util.GetParentDirectory(oldPath, 1)
		c.walPath = oldPath + "/data"
		c.cfgPath = oldPath + "/config"
		c.oBlockDb = dbm.NewDB("blockstore", dbm.LevelDBBackend, oldPath+"/data")
		c.oStateDb = dbm.NewDB("state", dbm.LevelDBBackend, oldPath+"/data")
		c.oEvidenceDb = dbm.NewDB("evidence", dbm.LevelDBBackend, oldPath+"/data")
		c.oTrustDb = dbm.NewDB("trusthistory", dbm.GoLevelDBBackend, oldPath+"/data")
	}

	if newPath != "" {
		c.ethDbPath = util.GetParentDirectory(newPath, 1)
		c.walPath = newPath + "/data"
		c.cfgPath = newPath + "/config"
		c.nBlockDb = db.NewDB("blockstore", db.LevelDBBackend, newPath+"/data")
		c.nStateDb = db.NewDB("state", db.LevelDBBackend, newPath+"/data")
		c.nEvidenceDb = db.NewDB("evidence", db.LevelDBBackend, newPath+"/data")
		c.nTrustDb = db.NewDB("trusthistory", db.GoLevelDBBackend, newPath+"/data")
	}
}

func (c *TmDataStore) OnStop() {
	c.close1(c.oBlockDb)
	c.close1(c.oStateDb)
	c.close1(c.oEvidenceDb)
	c.close1(c.oTrustDb)

	c.close2(c.nBlockDb)
	c.close2(c.nStateDb)
	c.close2(c.nEvidenceDb)
	c.close2(c.nTrustDb)
}

func (c *TmDataStore) close1(ldb dbm.DB) {
	if ldb != nil {
		ldb.Close()
	}
}

func (c *TmDataStore) close2(ldb db.DB) {
	if ldb != nil {
		ldb.Close()
	}
}

func (c *TmDataStore) OnGenesisJSON(oPath, nPath string) {
	cvt.UpgradeGenesisJSON(oPath, nPath)

	genDoc, err := types.GenesisDocFromFile(nPath)
	if err == nil {
		util.SaveNewGenesisDoc(c.nStateDb, genDoc)
	}
}

func (c *TmDataStore) OnPrivValidatorJSON(oPath, nPath string) {
	privVali := privval.GenFilePV(nPath)
	cvt.NewPrivValidator(oPath, privVali)
	privVali.Save()
}

func (c *TmDataStore) TotalHeight(isNew bool) {
	if isNew {
		c.totalHeight = util.LoadNewTotalHeight(c.nBlockDb)
	} else {
		c.totalHeight = cvt.LoadOldTotalHeight(c.oBlockDb)
	}
	fmt.Println("total height", c.totalHeight)
}

func (c *TmDataStore) OnBlockStore(startHeight int64) {
	if startHeight < 1 {
		cmn.Exit(fmt.Sprint("Invalid start height"))
	}

	c.newState = cvt.InitState(c.oStateDb)

	var lastBlockID *types.BlockID
	if startHeight == 1 {
		lastBlockID = &types.BlockID{}
	} else {
		nMeta := util.LoadNewBlockMeta(c.nBlockDb, startHeight-1)
		lastBlockID = &nMeta.BlockID
	}

	cnt := 0
	limit := 1000
	index := startHeight
	batch := c.nBlockDb.NewBatch()
	for ; index <= c.totalHeight; index++ {
		cnt++

		nBlock := cvt.NewBlockFromOld(c.oBlockDb, index, lastBlockID, c.newState)
		seenCommit := cvt.NewSeenCommit(c.oBlockDb, index, c.newState)

		blockParts := nBlock.MakePartSet(c.newState.ConsensusParams.BlockPartSizeBytes)
		nMeta := types.NewBlockMeta(nBlock, blockParts)

		c.saveBlock(batch, nBlock, nMeta, seenCommit)
		if cnt%limit == 0 {
			log.Printf("batch write %v/%v\n", cnt, c.totalHeight)
			batch.WriteSync()
			batch = c.nBlockDb.NewBatch()
		}

		// update lastBlockID
		lastBlockID = &nMeta.BlockID

		// upgrade state data
		c.upgradeStateData(index)
	}
	if cnt%limit != 0 {
		log.Printf("batch write %v/%v\n", cnt, c.totalHeight)
		batch.WriteSync()
	}
	c.upgradeStateData(index)

	c.newState.LastBlockID = *lastBlockID
	util.SaveNewState(c.nStateDb, c.newState)
}

func (c *TmDataStore) upgradeStateData(height int64) {
	cvt.SaveNewABCIResponse(c.oStateDb, c.nStateDb, height)
	cvt.SaveNewConsensusParams(c.oStateDb, c.nStateDb, height)
	cvt.SaveNewValidators(c.oStateDb, c.nStateDb, height)
}

func (c *TmDataStore) saveBlock(batch dbm.Batch, block *types.Block, blockMeta *types.BlockMeta, seenCommit *types.Commit) {
	height := block.Height
	util.SaveNewBlockMeta2(batch, height, blockMeta)
	util.SaveNewBlockParts2(batch, height, block, c.newState)
	util.SaveNewCommit2(batch, height-1, "C", block.LastCommit)
	util.SaveNewCommit2(batch, height, "SC", seenCommit)
	util.SaveNewBlockStoreStateJSON2(batch, height)
}

func (c *TmDataStore) OnEvidence() {
	cvt.UpgradeEvidence(c.oEvidenceDb, c.nEvidenceDb)
}

func (c *TmDataStore) OnBlockRecover(newVersion bool, resetHeight int64) {
	if c.totalHeight <= resetHeight {
		cmn.Exit(fmt.Sprintf("reset height %d >= total height %d", resetHeight, c.totalHeight))
	}
	c.resetEthBlock(resetHeight)

	wal := CreateTmWal(c.walPath)
	prival := CreateTmCfgPrival(c.cfgPath)
	if newVersion {
		c.resetNewBlock(resetHeight)
		wal.ResetNewWalHeight(resetHeight)
		prival.ResetNewPVHeight(resetHeight)
	} else {
		c.resetOldBlock(resetHeight)
		wal.ResetOldWalHeight(resetHeight)
		prival.ResetOldPVHeight(resetHeight)
	}
}

func (c *TmDataStore) resetEthBlock(height int64) {
	eth := CreateEthDb(c.ethDbPath)
	defer eth.OnStop()

	eth.ResetBlockHeight(height)
}

func (c *TmDataStore) OnEvidenceRecover(newVersion bool, resetHeight int64) {

}

func (c *TmDataStore) resetOldBlock(height int64) {
	state := c.getOldState(height)
	util.SaveOldState(c.oStateDb, state)

	for i := height + 1; i <= c.totalHeight; i++ {
		block := cvt.LoadOldBlock(c.oBlockDb, i)
		c.deleteOldBlock(block, state)
	}
	json := his.BlockStoreStateJSON{Height: height}
	cvt.SaveOldBlockStoreStateJson(c.oBlockDb, json)
}

func (c *TmDataStore) resetNewBlock(height int64) {
	state := c.getNewState(height)
	util.SaveNewState(c.nStateDb, state)

	for i := height + 1; i <= c.totalHeight; i++ {
		block := util.LoadNewBlock(c.nBlockDb, i)
		c.deleteNewBlock(block, state)
	}
	util.SaveNewBlockStoreStateJSON(c.nBlockDb, height)
}

func (c *TmDataStore) getOldState(height int64) *his.State {
	state := new(his.State)

	oState := cvt.LoadOldState(c.oStateDb)
	lastBlock := cvt.LoadOldBlock(c.oBlockDb, height-1)
	block := cvt.LoadOldBlock(c.oBlockDb, height)

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

func (c *TmDataStore) getNewState(height int64) *state.State {
	state := new(state.State)

	oState := util.LoadNewState(c.nStateDb)
	lastBlock := util.LoadNewBlock(c.nBlockDb, height-1)
	block := util.LoadNewBlock(c.nBlockDb, height)

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

func (c *TmDataStore) deleteOldBlock(block *his.Block, state *his.State) {
	// block
	util.DeleteBlockMeta(false, c.oBlockDb, c.nBlockDb, block.Height)
	util.DeleteOldBlockParts(c.oBlockDb, block, state)
	util.DeleteCommit(false, c.oBlockDb, c.nBlockDb, block.Height)

	// state
	stateHeight := block.Height + 1
	cvt.DeleteABCIResponse(false, c.oStateDb, c.nStateDb, block.Height)
	cvt.DeleteConsensusParam(false, c.oStateDb, c.nStateDb, stateHeight)
	cvt.DeleteValidator(false, c.oStateDb, c.nStateDb, stateHeight)
}

func (c *TmDataStore) deleteNewBlock(block *types.Block, state *state.State) {
	// block
	util.DeleteBlockMeta(true, c.oBlockDb, c.nBlockDb, block.Height)
	util.DeleteNewBlockParts(c.nBlockDb, block, state)
	util.DeleteCommit(true, c.oBlockDb, c.nBlockDb, block.Height)

	// state
	stateHeight := block.Height + 1
	cvt.DeleteABCIResponse(true, c.oStateDb, c.nStateDb, block.Height)
	cvt.DeleteConsensusParam(true, c.oStateDb, c.nStateDb, stateHeight)
	cvt.DeleteValidator(true, c.oStateDb, c.nStateDb, stateHeight)
}

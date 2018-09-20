package convert

import (
	"fmt"
	"log"

	"github.com/commis/tm-tools/libs/util"
	"github.com/commis/tm-tools/oldver"
	oldcvt "github.com/commis/tm-tools/oldver/convert"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

type Converter struct {
	oBlockDb, oStateDb, oEvidenceDb, oTrustDb dbm.DB
	nBlockDb, nStateDb, nEvidenceDb, nTrustDb db.DB
	totalHeight                               int64
	newState                                  *state.State
}

func (c *Converter) OnStart(oldPath, newPath string) {
	c.oBlockDb = dbm.NewDB("blockstore", dbm.LevelDBBackend, oldPath+"/data")
	c.oStateDb = dbm.NewDB("state", dbm.LevelDBBackend, oldPath+"/data")
	c.oEvidenceDb = dbm.NewDB("evidence", dbm.LevelDBBackend, oldPath+"/data")
	c.oTrustDb = dbm.NewDB("trusthistory", dbm.GoLevelDBBackend, oldPath+"/data")

	c.nBlockDb = db.NewDB("blockstore", db.LevelDBBackend, newPath+"/data")
	c.nStateDb = db.NewDB("state", db.LevelDBBackend, newPath+"/data")
	c.nEvidenceDb = db.NewDB("evidence", db.LevelDBBackend, newPath+"/data")
	c.nTrustDb = db.NewDB("trusthistory", db.GoLevelDBBackend, newPath+"/data")
}

func (c *Converter) OnStop() {
	c.oBlockDb.Close()
	c.oStateDb.Close()
	c.oEvidenceDb.Close()
	c.oTrustDb.Close()

	c.nBlockDb.Close()
	c.nStateDb.Close()
	c.nEvidenceDb.Close()
	c.nTrustDb.Close()
}

func (c *Converter) OnGenesisJSON(oPath, nPath string) {
	oldcvt.UpgradeGenesisJSON(c.nStateDb, oPath, nPath)
}

func (c *Converter) OnPrivValidatorJSON(oPath, nPath string) {
	privVali := privval.GenFilePV(nPath)
	oldcvt.NewPrivValidator(oPath, privVali)
	privVali.Save()
}

func (c *Converter) TotalHeight() {
	c.totalHeight = oldver.LoadTotalHeight(c.oStateDb)
	fmt.Println("total height", c.totalHeight)
}

func (c *Converter) OnBlockStore(startHeight int64) {
	if startHeight < 1 {
		panic("Invalid start height")
	}

	c.newState = oldcvt.InitState(c.oStateDb)

	var lastBlockID *types.BlockID
	if startHeight == 1 {
		lastBlockID = &types.BlockID{}
	} else {
		nMeta := util.LoadNewBlockMeta(c.nBlockDb, startHeight-1)
		lastBlockID = &nMeta.BlockID
	}

	cnt := 0
	limit := 1000
	batch := c.nBlockDb.NewBatch()
	for i := startHeight; i <= c.totalHeight; i++ {
		cnt++

		nBlock := oldcvt.NewBlockFromOld(c.oBlockDb, i, lastBlockID, c.newState)
		blockParts := nBlock.MakePartSet(c.newState.ConsensusParams.BlockPartSizeBytes)
		nMeta := types.NewBlockMeta(nBlock, blockParts)
		// seen this BlockId's commit
		seenCommit := oldcvt.NewSeenCommit(c.oBlockDb, i, &nMeta.BlockID, c.newState)

		c.saveBlock(batch, nBlock, nMeta, seenCommit)
		if cnt%limit == 0 {
			log.Printf("batch write %v/%v\n", cnt, c.totalHeight)
			batch.Write()
			batch = c.nBlockDb.NewBatch()
		}

		// update lastBlockID
		lastBlockID = &nMeta.BlockID

		// upgrade state data
		//oldver.SaveNewABCIResponses(oStateDb, nStateDb, nMeta.Header.Height)
		oldver.SaveNewConsensusParams(c.oStateDb, c.nStateDb, nMeta.Header.Height)
		oldver.SaveNewValidators(c.oStateDb, c.nStateDb, nMeta.Header.Height)
	}
	if cnt%limit != 0 {
		log.Printf("batch write %v/%v\n", cnt, c.totalHeight)
		batch.Write()
	}

	c.saveState(lastBlockID)
}

func (c *Converter) saveBlock(batch dbm.Batch, block *types.Block, blockMeta *types.BlockMeta, seenCommit *types.Commit) {
	height := block.Height
	util.SaveNewBlockMeta2(batch, height, blockMeta)
	util.SaveNewBlockParts2(batch, height, block, c.newState)
	util.SaveNewCommit2(batch, height-1, "C", block.LastCommit)
	util.SaveNewCommit2(batch, height, "SC", seenCommit)
	util.SaveNewBlockStoreStateJSON2(batch, height)
}

func (c *Converter) saveState(lastBlockID *types.BlockID) {
	c.newState.LastBlockID = *lastBlockID
	util.SaveNewState(c.nStateDb, c.newState)
}

func (c *Converter) OnEvidence() {
	oldcvt.UpgradeEvidence(c.oEvidenceDb, c.nEvidenceDb)
}

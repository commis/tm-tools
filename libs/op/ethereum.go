package op

import (
	"github.com/commis/tm-tools/libs/log"
	"github.com/commis/tm-tools/libs/util"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	cmn "github.com/tendermint/tendermint/libs/common"
)

func ResetEthHeight(dbPath string, height int64) {
	tmEth := CreateTmEthDb(dbPath)
	defer tmEth.Close()

	tmEth.ResetHeight(height)
}

type TmEthDb struct {
	dbPath string
	db     *ethdb.LDBDatabase
}

func CreateTmEthDb(tmPath string) *TmEthDb {
	chainPath := util.GetParentDir(tmPath, 1)

	tmEth := &TmEthDb{
		dbPath: chainPath + "/ethermint/chaindata",
	}

	dbh, err := ethdb.NewLDBDatabase(tmEth.dbPath, 0, 256)
	if err != nil {
		cmn.Exit(err.Error())
	}

	tmEth.db = dbh
	return tmEth
}

func (te *TmEthDb) ResetHeight(height int64) {
	endNumber := uint64(height)
	totalHeight := te.getLastHeader()

	for i := totalHeight; i > endNumber; i-- {
		hash := core.GetCanonicalHash(te.db, i)
		core.DeleteBlock(te.db, hash, i)
	}
	te.setHeader(endNumber)
	te.getLastHeader()
}

func (te *TmEthDb) Close() {
	if te.db != nil {
		te.db.Close()
	}
}

func (te *TmEthDb) GetAll() {
	it := te.db.NewIterator()
	it.Seek([]byte(""))
	defer func() {
		if it != nil {
			it.Release()
		}
	}()

	for it.Next() {
		key := it.Key()
		log.Infof("%s\n", string(key))
	}
}

func (te *TmEthDb) PrintHeader() {
	hash := core.GetHeadBlockHash(te.db)
	number := core.GetBlockNumber(te.db, hash)
	log.Infof("print header number: %d ===\n", number)
	log.Infof("  %s\n", core.GetHeadHeaderHash(te.db).String())    //headHeaderKey
	log.Infof("  %s\n", core.GetHeadBlockHash(te.db).String())     //headBlockKey
	log.Infof("  %s\n", core.GetHeadFastBlockHash(te.db).String()) //headFastKey
}

func (te *TmEthDb) getLastHeader() uint64 {
	hash := core.GetHeadBlockHash(te.db)
	number := core.GetBlockNumber(te.db, hash)
	log.Infof("header: %d %s\n", number, hash.String())

	return number
}

func (te *TmEthDb) setHeader(height uint64) {
	var err error
	hash := core.GetCanonicalHash(te.db, height)
	if err = core.WriteCanonicalHash(te.db, hash, height); err != nil {
		log.Errorf("write canonical hash failed.%d %s\n", height, hash.String())
	}

	if err = core.WriteHeadHeaderHash(te.db, hash); err != nil {
		log.Errorf("write head header hash failed.%d %s\n", height, hash.String())
	}

	if err = core.WriteHeadBlockHash(te.db, hash); err != nil {
		log.Errorf("write head block hash failed.%d %s\n", height, hash.String())
	}

	if err = core.WriteHeadFastBlockHash(te.db, hash); err != nil {
		log.Errorf("write canonical hash failed.%d %s\n", height, hash.String())
	}
}

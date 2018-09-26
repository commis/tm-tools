package op

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type EthDB struct {
	db *ethdb.LDBDatabase
}

func CreateEthDb(dbPath string) *EthDB {
	eth := &EthDB{}
	var err error
	eth.db, err = ethdb.NewLDBDatabase(dbPath+"/ethermint/chaindata", 0, 256)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return eth
}

func (e *EthDB) OnStop() {
	if e.db != nil {
		e.db.Close()
	}
}

func (e *EthDB) GetAll() {
	it := e.db.NewIterator()
	it.Seek([]byte(""))
	defer func() {
		if it != nil {
			it.Release()
		}
	}()

	for it.Next() {
		key := it.Key()
		fmt.Printf("%s\n", string(key))
	}
}

func (e *EthDB) ResetBlockHeight(height int64) {
	endNumber := uint64(height)
	totalHeight := e.getLastHeader()

	for i := totalHeight; i > endNumber; i-- {
		hash := core.GetCanonicalHash(e.db, i)
		core.DeleteBlock(e.db, hash, i)
	}
	e.setHeader(endNumber)
	e.getLastHeader()
}

func (e *EthDB) getLastHeader() uint64 {
	hash := core.GetHeadBlockHash(e.db)
	number := core.GetBlockNumber(e.db, hash)
	fmt.Printf("header: %d %s\n", number, hash.String())
	return number
}

func (e *EthDB) printHeader() {
	hash := core.GetHeadBlockHash(e.db)
	number := core.GetBlockNumber(e.db, hash)
	fmt.Printf("print header number: %d ===\n", number)
	fmt.Printf("  %s\n", core.GetHeadHeaderHash(e.db).String())    //headHeaderKey
	fmt.Printf("  %s\n", core.GetHeadBlockHash(e.db).String())     //headBlockKey
	fmt.Printf("  %s\n", core.GetHeadFastBlockHash(e.db).String()) //headFastKey
}

func (e *EthDB) setHeader(height uint64) {
	var err error
	hash := core.GetCanonicalHash(e.db, height)
	if err = core.WriteCanonicalHash(e.db, hash, height); err != nil {
		fmt.Printf("write canonical hash failed.%d %s\n", height, hash.String())
	}

	if err = core.WriteHeadHeaderHash(e.db, hash); err != nil {
		fmt.Printf("write head header hash failed.%d %s\n", height, hash.String())
	}

	if err = core.WriteHeadBlockHash(e.db, hash); err != nil {
		fmt.Printf("write head block hash failed.%d %s\n", height, hash.String())
	}

	if err = core.WriteHeadFastBlockHash(e.db, hash); err != nil {
		fmt.Printf("write canonical hash failed.%d %s\n", height, hash.String())
	}
}

package commands

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/commis/tm-tools/libs/log"

	"github.com/commis/tm-tools/libs/util"
	cvt "github.com/commis/tm-tools/oldver/convert"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tmlibs/db"
)

type DbType string

type ViewParam struct {
	dbPath  string
	action  string
	key     string
	ver     bool
	height int64
	limit   int64
	decode  bool
}

var vp = &ViewParam{}

func init() {
	ViewDatabaseCmd.Flags().StringVar(&vp.dbPath, "db", "/home/tendermint/data/blockstore",
		"Database full path for the viewer")
	ViewDatabaseCmd.Flags().StringVar(&vp.action, "a", "get", "Operate database for [get|getall|block]")
	ViewDatabaseCmd.Flags().StringVar(&vp.key, "q", "", "Database query key")
	ViewDatabaseCmd.Flags().BoolVar(&vp.ver, "v", false, "Whether new version data")
	ViewDatabaseCmd.Flags().Int64Var(&vp.height, "h", 1, "View the block height")
	ViewDatabaseCmd.Flags().Int64Var(&vp.limit, "l", 0, "Limit of query list")
	ViewDatabaseCmd.Flags().BoolVar(&vp.decode, "d", false, "Whether decode data")
}

var ViewDatabaseCmd = &cobra.Command{
	Use:   "view",
	Short: "Tendermint database viewer",
	RunE:  viewDatabase,
}

func viewDatabase(cmd *cobra.Command, args []string) error {
	holder := CreateViewDbHolder(vp.ver, vp.dbPath)
	defer holder.Close()

	switch vp.action {
	case "get":
		holder.GetDataByKey()
		break
	case "getall":
		holder.GetAllRecordKeys()
		break
	case "block":
		holder.GetBlock()
		break
	default:
		cmn.Exit(fmt.Sprintf("action is invalid '%s'", vp.action))
	}

	return nil
}

type DbHandler struct {
	LevelDb *leveldb.DB
	OldDbm  dbm.DB
	NewDbm  db.DB
}

func CreateViewDbHolder(newVersion bool, path string) *DbHandler {
	dbName := util.FileNameNoExt(path)
	dbPath := filepath.Dir(path)

	handler := new(DbHandler)
	if newVersion {
		ldb := db.NewDB(dbName, db.LevelDBBackend, dbPath)
		handler.LevelDb = ldb.(*db.GoLevelDB).DB()
		handler.NewDbm = ldb
	} else {
		ldb := dbm.NewDB(dbName, dbm.LevelDBBackend, dbPath)
		handler.LevelDb = ldb.(*dbm.GoLevelDB).DB()
		handler.OldDbm = ldb
	}

	return handler
}

func (d *DbHandler) Close() {
	d.LevelDb.Close()
}

func (d *DbHandler) GetDataByKey() {
	data := d.getData(vp.key)
	if len(data) == 0 {
		log.Errorf("viewer data %s is not exist", vp.key)
		return
	}

	if !vp.decode {
		fmt.Println(string(data))
		return
	}

	switch vp.key {
	case util.StateKey:
		d.loadState()
		break
	case util.GenesisDoc:
		d.loadGenesisDoc()
		break
	case util.BlockStoreKey:
		d.loadBlockStore()
		break
	default:
		p := strings.Split(vp.key, ":")
		if len(p) >= 2 {
			height, _ := strconv.ParseInt(p[1], 10, 64)
			firstKey := p[0]
			switch firstKey {
			case "H":
				d.loadBlockMeta(height)
				break
			case "P":
				index, _ := strconv.Atoi(p[2])
				d.loadBlockPart(height, index)
				break
			case "C":
				d.loadBlockCommit(height, firstKey)
				break
			case "SC":
				d.loadBlockCommit(height, firstKey)
				break
			case util.ABCIResponsesKey:
				d.loadABCIResponse(height)
				break
			case util.ConsensusParamsKey:
				d.loadConsensusParam(height)
				break
			case util.ValidatorsKey:
				d.loadValidator(height)
				break
			default:
				fmt.Println(string(data))
			}
		}
	}
}

func (d *DbHandler) getData(key string) []byte {
	if key == "" {
		return []byte{}
	}

	res, err := d.LevelDb.Get([]byte(key), nil)
	if err != nil {
		log.Debugf("failed to get data by key %v", err)
		return []byte{}
	}
	return res
}

func (d *DbHandler) loadState() {
	var res []byte

	if d.OldDbm != nil {
		state := cvt.LoadOldState(d.OldDbm)
		res, _ = json.Marshal(state)
	} else if d.NewDbm != nil {
		state := util.LoadNewState(d.NewDbm)
		res, _ = json.Marshal(state)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadGenesisDoc() {
	var res []byte

	if d.OldDbm != nil {
		genDoc := cvt.LoadOldGenesisDoc(d.OldDbm)
		res, _ = json.Marshal(genDoc)
	} else if d.NewDbm != nil {
		genDoc := util.LoadNewGenesisDoc(d.NewDbm)
		res, _ = json.Marshal(genDoc)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadBlockStore() {
	var res []byte

	if d.OldDbm != nil {
		blockstore := cvt.LoadOldBlockStoreStateJSON(d.OldDbm)
		res, _ = json.Marshal(blockstore)
	} else if d.NewDbm != nil {
		blockstore := util.LoadNewBlockStoreStateJSON(d.NewDbm)
		res, _ = json.Marshal(blockstore)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadBlockMeta(height int64) {
	var res []byte

	if d.OldDbm != nil {
		meta := cvt.LoadOldBlockMeta(d.OldDbm, height)
		res, _ = json.Marshal(meta)
	} else if d.NewDbm != nil {
		meta := util.LoadNewBlockMeta(d.NewDbm, height)
		res, _ = json.Marshal(meta)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadBlockPart(height int64, index int) {
	var res []byte

	if d.OldDbm != nil {
		part := cvt.LoadOldBlockPart(d.OldDbm, height, index)
		res, _ = json.Marshal(part)
	} else if d.NewDbm != nil {
		part := util.LoadNewBlockPart(d.NewDbm, height, index)
		res, _ = json.Marshal(part)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadBlockCommit(height int64, prefix string) {
	var res []byte

	if d.OldDbm != nil {
		commit := cvt.LoadOldBlockCommit(d.OldDbm, height, prefix)
		res, _ = json.Marshal(commit)
	} else if d.NewDbm != nil {
		commit := util.LoadNewBlockCommit(d.NewDbm, height, prefix)
		res, _ = json.Marshal(commit)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadABCIResponse(height int64) {
	var res []byte

	if d.OldDbm != nil {
		response := cvt.LoadOldABCIResponse(d.OldDbm, height)
		if response != nil {
			res, _ = json.Marshal(response)
		}
	} else if d.NewDbm != nil {
		response := util.LoadNewABCIResponse(d.NewDbm, height)
		if response != nil {
			res, _ = json.Marshal(response)
		}
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadConsensusParam(height int64) {
	var res []byte

	if d.OldDbm != nil {
		consensus := cvt.LoadOldConsensusParamsInfo(d.OldDbm, height)
		if consensus != nil {
			res, _ = json.Marshal(consensus)
		}
	} else if d.NewDbm != nil {
		consensus := util.LoadNewConsensusParamsInfo(d.NewDbm, height)
		if consensus != nil {
			res, _ = json.Marshal(consensus)
		}
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadValidator(height int64) {
	var res []byte

	if d.OldDbm != nil {
		validator := cvt.LoadOldValidatorsInfo(d.OldDbm, height)
		res, _ = json.Marshal(validator)
	} else if d.NewDbm != nil {
		validator := util.LoadNewValidatorsInfo(d.NewDbm, height)
		res, _ = json.Marshal(validator)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) GetAllRecordKeys() {
	query := d.LevelDb.NewIterator(nil, nil)
	defer query.Release()

	var index int64 = 1
	query.Seek([]byte(vp.key))
	for query.Next() {
		if vp.limit != 0 && index%vp.limit == 0 {
			break
		}
		index++
		fmt.Printf("%s\n", string(query.Key()))
	}
}

func (d *DbHandler) GetBlock() {
	var res []byte

	if d.OldDbm != nil {
		block := cvt.LoadOldBlock(d.OldDbm, vp.height)
		res, _ = json.Marshal(block)
	} else if d.NewDbm != nil {
		block := util.LoadNewBlock(d.NewDbm, vp.height)
		res, _ = json.Marshal(block)
	}
	fmt.Println(string(res))
}

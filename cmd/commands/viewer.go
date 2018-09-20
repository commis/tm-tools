package commands

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/commis/tm-tools/libs/util"
	oldver "github.com/commis/tm-tools/oldver"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	dbPath  string
	action  string
	key     string
	ver     string
	iHeight int64
	limit   int64
	decode  bool
)

func init() {
	ViewDatabaseCmd.Flags().StringVar(&dbPath, "db", "/home/tendermint/data/blockstore",
		"Database full path for the viewer")
	ViewDatabaseCmd.Flags().StringVar(&action, "a", "get", "Operate database for [get|getall|block]")
	ViewDatabaseCmd.Flags().StringVar(&key, "q", "", "Database query key")
	ViewDatabaseCmd.Flags().StringVar(&ver, "v", "new", "Database version of [new|oldver]")
	ViewDatabaseCmd.Flags().Int64Var(&iHeight, "h", 1, "View the block height")
	ViewDatabaseCmd.Flags().Int64Var(&limit, "l", 0, "Limit of query list")
	ViewDatabaseCmd.Flags().Bool("d", decode, "Whether decode data")
}

var ViewDatabaseCmd = &cobra.Command{
	Use:   "view",
	Short: "Tendermint database viewer",
	RunE:  viewDatabase,
}

func viewDatabase(cmd *cobra.Command, args []string) error {
	holder := CreateViewDbHolder(ver, dbPath)
	defer holder.Close()

	switch action {
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
		panic(fmt.Sprintf("action is invalid '%s'", action))
	}

	return nil
}

type DbHandler struct {
	LevelDb *leveldb.DB
	OldDbm  dbm.DB
	NewDbm  db.DB
}

func CreateViewDbHolder(ver, path string) *DbHandler {
	dbName := util.FileNameNoExt(path)
	dbPath := filepath.Dir(path)

	handler := new(DbHandler)
	switch ver {
	case "oldver":
		ldb := dbm.NewDB(dbName, dbm.LevelDBBackend, dbPath)
		handler.LevelDb = ldb.(*dbm.GoLevelDB).DB()
		handler.OldDbm = ldb
		break
	case "new":
		ldb := db.NewDB(dbName, db.LevelDBBackend, dbPath)
		handler.LevelDb = ldb.(*db.GoLevelDB).DB()
		handler.NewDbm = ldb
		break
	default:
		panic(fmt.Sprintf("version is invalid '%s'", ver))
	}

	return handler
}

func (d *DbHandler) Close() {
	d.LevelDb.Close()
}

func (d *DbHandler) GetDataByKey() {
	data := d.getData(key)
	if len(data) == 0 {
		fmt.Println(key, "not exist")
		return
	}

	if !decode {
		fmt.Println(string(data))
		return
	}

	switch key {
	case util.StateKey:
		break
	case util.ABCIResponsesKey:
		break
	default:
		p := strings.Split(key, ":")
		if len(p) >= 2 {
			height, _ := strconv.ParseInt(p[1], 10, 64)
			switch p[0] {
			case "H":
				d.loadBlockMeta(height)
				break
			case "P":
				index, _ := strconv.Atoi(p[2])
				d.loadBlockPart(height, index)
				break
			case "C":
			case "SC":
				d.loadBlockCommit(height, p[0])
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
		panic(err)
	}
	return res
}

func (d *DbHandler) loadBlockMeta(height int64) {
	var res []byte

	if d.OldDbm != nil {
		meta := oldver.LoadOldState(d.OldDbm)
		res, _ = json.Marshal(meta)
	} else if d.NewDbm != nil {
		meta := util.LoadNewState(d.NewDbm)
		res, _ = json.Marshal(meta)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) loadBlockPart(height int64, index int) {
	var res []byte

	if d.OldDbm != nil {
		part := oldver.LoadOldBlockPart(d.OldDbm, height, index)
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
		commit := oldver.LoadOldBlockCommit(d.OldDbm, height, prefix)
		res, _ = json.Marshal(commit)
	} else if d.NewDbm != nil {
		commit := util.LoadNewBlockCommit(d.NewDbm, height, prefix)
		res, _ = json.Marshal(commit)
	}
	fmt.Println(string(res))
}

func (d *DbHandler) GetAllRecordKeys() {
	query := d.LevelDb.NewIterator(nil, nil)
	defer query.Release()

	query.Seek([]byte(key))
	for query.Next() {
		fmt.Printf("%s\n", string(query.Key()))
	}
}

func (d *DbHandler) GetBlock() {
	var res []byte

	if d.OldDbm != nil {
		block := oldver.LoadOldBlock(d.OldDbm, iHeight)
		res, _ = json.Marshal(block)
	} else if d.NewDbm != nil {
		block := util.LoadNewBlock(d.NewDbm, iHeight)
		res, _ = json.Marshal(block)
	}
	fmt.Println(string(res))
}

package main

import (
	"fmt"
	"os"

	"github.com/commis/tm-tools/libs/log"

	"github.com/commis/tm-tools/libs/op"
)

type TmRepair struct {
	path  string
	isNew bool
	store *op.TmDataStore
}

func CreateTmRepair(dataPath string, isNew bool) *TmRepair {
	repair := &TmRepair{path: dataPath, isNew: isNew}
	repair.store = &op.TmDataStore{}
	if isNew {
		repair.store.OnStart("", dataPath)
	} else {
		repair.store.OnStart(dataPath, "")
	}

	return repair
}

func (t *TmRepair) Close() {
	if t.store != nil {
		t.store.OnStop()
	}
}

func (t *TmRepair) GetRepairFilePV() {
	t.store.TotalHeight(t.isNew)
	log.Infof("current block height: %d", t.store.Height)

	wal := op.CreateTmWal(t.path + "/data")
	if t.isNew {
		wal.ResetNewWalHeight(t.store.Height)
	} else {
		wal.ResetOldWalHeight(t.store.Height)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("missing two argument: <path-to-tendermint> <old|new>")
		os.Exit(1)
	}

	newVer := true
	if os.Args[2] == "old" {
		newVer = false
	}

	tm := CreateTmRepair(os.Args[1], newVer)
	defer tm.Close()

	tm.GetRepairFilePV()
}

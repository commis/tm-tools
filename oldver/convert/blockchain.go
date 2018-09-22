package convert

import (
	"encoding/json"
	"fmt"

	"github.com/commis/tm-tools/libs/util"
	oldtype "github.com/commis/tm-tools/oldver/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func LoadOldTotalHeight(ldb dbm.DB) int64 {
	blockStore := LoadOldBlockStoreStateJSON(ldb)
	return blockStore.Height
}

func LoadOldBlockStoreStateJSON(ldb dbm.DB) oldtype.BlockStoreStateJSON {
	bytes := ldb.Get([]byte(util.BlockStoreKey))
	if bytes == nil {
		return oldtype.BlockStoreStateJSON{
			Height: 0,
		}
	}
	bsj := oldtype.BlockStoreStateJSON{}
	err := json.Unmarshal(bytes, &bsj)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Could not unmarshal bytes: %X", bytes))
	}
	return bsj
}

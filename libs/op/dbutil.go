package op

import (
	"github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tmlibs/db"
)

type TMVersionType string

const (
	TMVer0180 TMVersionType = "0.18.0"
	TMVer0231 TMVersionType = "0.23.1"
)

func CloseDbm(ldb dbm.DB) {
	if ldb != nil {
		ldb.Close()
	}
}

func CloseDb(ldb db.DB) {
	if ldb != nil {
		ldb.Close()
	}
}

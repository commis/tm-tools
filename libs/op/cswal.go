package op

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/commis/tm-tools/libs/log"
	"github.com/commis/tm-tools/libs/op/hold"
	"github.com/commis/tm-tools/libs/util"
	ocs "github.com/commis/tm-tools/oldver/consensus"
	cvt "github.com/commis/tm-tools/oldver/convert"
	"github.com/tendermint/tendermint/consensus"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tmlibs/db"
)

func UpgradeNodeCsWal(oTmPath, nTmPath string) {
	tmWal := CreateTmCsWal(oTmPath, nTmPath)
	defer func() {
		CloseDbm(tmWal.oBlockDb)
		CloseDb(tmWal.nBlockDb)
	}()

	tmWal.upgradeMessage()
}

func createTmCsWalByVersion(ver TMVersionType, tmPath string) *TmCsWal {
	var tmWal *TmCsWal = nil
	switch ver {
	case TMVer0180:
		tmWal = CreateTmCsWal(tmPath, "")
	case TMVer0231:
		tmWal = CreateTmCsWal("", tmPath)
	}
	return tmWal
}

func ResetNodeWalHeight(ver TMVersionType, tmPath string, height int64) {
	tmWal := createTmCsWalByVersion(ver, tmPath)
	if tmWal != nil {
		defer func() {
			CloseDbm(tmWal.oBlockDb)
			CloseDb(tmWal.nBlockDb)
		}()
		tmWal.ResetHeight(height)
	}
}

func PrintWalMessage(ver TMVersionType, tmPath string) {
	tmWal := createTmCsWalByVersion(ver, tmPath)
	if tmWal != nil {
		defer func() {
			CloseDbm(tmWal.oBlockDb)
			CloseDb(tmWal.nBlockDb)
		}()
		tmWal.PrintWalMessage()
	}
}

func InitWalMessage(ver TMVersionType, tmPath string) {
	tmWal := createTmCsWalByVersion(ver, tmPath)
	if tmWal != nil {
		defer func() {
			CloseDbm(tmWal.oBlockDb)
			CloseDb(tmWal.nBlockDb)
		}()
		tmWal.InitWalMessage()
	}
}

type TmCsWal struct {
	tmPath  string
	oldPath string
	newPath string
	data    string
	memPool string

	oBlockDb dbm.DB
	nBlockDb db.DB
}

func CreateTmCsWal(oTmPath, nTmPath string) *TmCsWal {
	if oTmPath == "" && nTmPath == "" {
		cmn.Exit("create parameter is empty")
	}

	dt := "/data"
	wal := "/cs.wal"
	tmWal := &TmCsWal{
		data:    wal + "/wal",
		memPool: "/mempool.wal",
	}

	if oTmPath != "" {
		tmWal.tmPath = oTmPath
		tmWal.oldPath = oTmPath + dt
		tmWal.oBlockDb = dbm.NewDB("blockstore", dbm.LevelDBBackend, tmWal.oldPath)
	}

	if nTmPath != "" {
		tmWal.tmPath = nTmPath
		tmWal.newPath = nTmPath + dt
		util.CreateDirAll(tmWal.newPath + wal)
		tmWal.nBlockDb = db.NewDB("blockstore", db.LevelDBBackend, tmWal.newPath)
	}

	return tmWal
}

func (tm *TmCsWal) upgradeMessage() {
	oldWal := tm.getWalFile(tm.oldPath)
	newWal := tm.getWalFile(tm.newPath)

	rd, wd, err := tm.openWalFile(oldWal, newWal)
	if err != nil {
		return
	}
	defer tm.close(rd, wd)

	dec := ocs.NewWALDecoder(rd)
	enc := consensus.NewWALEncoderExt(wd)

	height := hold.LoadNewBlockHeight(tm.nBlockDb)
	bw := ocs.CreateBlockCsWal(tm.nBlockDb)
	for {
		msg, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to decode msg: %v", err)
			return
		}

		if !ocs.FilterWalMessage(height, msg.Msg) {
			continue
		}

		if walEvent := bw.ConvertWalMessage(msg); walEvent != nil {
			log.Debugf("event msg: %+v", *msg)
			if err := enc.Encode(walEvent); err != nil {
				log.Errorf("failed to encode msg: %v", err)
				return
			}
		}
	}
}

func (tm *TmCsWal) ResetHeight(height int64) {
	if tm.oldPath != "" {
		tm.resetOldWalHeight(height)
	} else if tm.newPath != "" {
		tm.resetNewWalHeight(height)
	}
}

func (tm *TmCsWal) PrintWalMessage() {
	if tm.oldPath != "" {
		tm.printOldWalMessage()
	} else if tm.newPath != "" {
		tm.printNewWalMessage()
	}
}

func (tm *TmCsWal) InitWalMessage() {
	if tm.oldPath != "" {
		height := cvt.LoadOldTotalHeight(tm.oBlockDb)
		ResetPrivValHeight(TMVer0180, tm.tmPath, height)
		tm.resetOldWalHeight(height)
	} else if tm.newPath != "" {
		height := hold.LoadNewBlockHeight(tm.nBlockDb)
		ResetPrivValHeight(TMVer0231, tm.tmPath, height)
		tm.resetNewWalHeight(height)
	}
}

func (tm *TmCsWal) resetOldWalHeight(height int64) {
	walFile := tm.getWalFile(tm.oldPath)
	tmpWalFile := tm.getWalFile(tm.oldPath) + ".tmp"

	rd, wd, err := tm.openWalFile(walFile, tmpWalFile)
	if err != nil {
		return
	}
	defer func() {
		util.Rename(walFile, tmpWalFile)
		os.RemoveAll(tm.oldPath + tm.memPool)
	}()

	lastHeight := height + 1
	dec := ocs.NewWALDecoder(rd)
	enc := ocs.NewWALEncoder(wd)
	bw := ocs.CreateBlockCsWal(tm.nBlockDb)
	for {
		msg, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to decode msg: %v", err)
			return
		}

		if ocs.FilterWalMessage(lastHeight, msg.Msg) {
			continue
		}

		bw.UpdateOldVoteToPrivVal(msg)
		if err := enc.Encode(msg); err != nil {
			log.Errorf("failed to encode msg: %v", err)
			return
		}
	}
}

func (tm *TmCsWal) resetNewWalHeight(height int64) {
	walFile := tm.getWalFile(tm.newPath)
	tmpWalFile := tm.getWalFile(tm.newPath) + ".tmp"

	rd, wd, err := tm.openWalFile(walFile, tmpWalFile)
	if err != nil {
		return
	}
	defer func() {
		util.Rename(walFile, tmpWalFile)
		os.RemoveAll(tm.newPath + tm.memPool)
	}()

	lastHeight := height + 1
	dec := consensus.NewWALDecoder(rd)
	enc := consensus.NewWALEncoderExt(wd)
	bw := ocs.CreateBlockCsWal(tm.nBlockDb)
	for {
		msg, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to decode msg: %v", err)
			return
		}

		if consensus.FilterBlockWalMessage(lastHeight, msg.Msg) {
			continue
		}

		bw.UpdateNewVoteToPrivVal(msg)
		if err := enc.Encode(msg); err != nil {
			log.Errorf("failed to encode msg: %v", err)
			return
		}
	}
}

func (tm *TmCsWal) printOldWalMessage() {
	walFile := tm.getWalFile(tm.oldPath)

	f, err := os.Open(walFile)
	if err != nil {
		log.Errorf("failed to open WAL file: %v", err)
	}
	defer f.Close()

	dec := ocs.NewWALDecoder(f)
	for {
		msg, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to decode msg: %v", err)
		}

		json, err := json.Marshal(msg)
		if err != nil {
			panic(fmt.Errorf("failed to marshal msg: %v", err))
		}

		fmt.Printf("%s\n", string(json))
		if end, ok := msg.Msg.(ocs.EndHeightMessage); ok {
			fmt.Printf("ENDHEIGHT %d\n", end.Height)
		}
	}
}

func (tm *TmCsWal) printNewWalMessage() {
	walFile := tm.getWalFile(tm.newPath)

	f, err := os.Open(walFile)
	if err != nil {
		log.Errorf("failed to open WAL file: %v", err)
	}
	defer f.Close()

	dec := consensus.NewWALDecoder(f)
	for {
		msg, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to decode msg: %v", err)
		}

		json, err := json.Marshal(msg)
		if err != nil {
			panic(fmt.Errorf("failed to marshal msg: %v", err))
		}

		fmt.Printf("%s\n", string(json))
		if end, ok := msg.Msg.(consensus.EndHeightMessage); ok {
			fmt.Printf("ENDHEIGHT %d\n", end.Height)
		}
	}
}

func (tm *TmCsWal) getWalFile(path string) string {
	return path + tm.data
}

func (tm *TmCsWal) openWalFile(old, new string) (rd *os.File, wd *os.File, err error) {
	rd, err = os.Open(old)
	if err == nil {
		wd, err = os.Create(new)
		if err != nil {
			rd.Close()
			log.Errorf("failed to create WAL file: %v", err)
		}
	} else {
		log.Errorf("failed to open WAL file: %v", err)
	}

	return
}

func (tm *TmCsWal) close(rd *os.File, wd *os.File) {
	if rd != nil {
		rd.Close()
	}

	if wd != nil {
		wd.Close()
	}
}

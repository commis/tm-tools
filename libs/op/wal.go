package op

import (
	"io"
	"os"

	"github.com/commis/tm-tools/libs/log"
	"github.com/commis/tm-tools/libs/util"
	cs "github.com/commis/tm-tools/oldver/consensus"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/consensus"
)

var cdc = amino.NewCodec()

func init() {
	consensus.RegisterConsensusMessages(cdc)
	consensus.RegisterWALMessages(cdc)
}

type TmWal struct {
	mempool string
	cswal   string
}

func CreateTmWal(walPath string) *TmWal {
	log.Infof("wal path: %s", walPath)
	return &TmWal{
		mempool: walPath + "/mempool.wal",
		cswal:   walPath + "/cs.wal/wal"}
}

func (t *TmWal) ResetOldWalHeight(height int64) {
	writeFile := t.cswal + ".new"
	rd, wd, err := t.open(writeFile)
	if err != nil {
		return
	}
	defer func() {
		t.close(rd)
		t.close(wd)
		util.Rename(t.cswal, writeFile)
		os.RemoveAll(t.mempool)
	}()

	lastHeight := height + 1
	dec := cs.NewWALDecoder(rd)
	enc := cs.NewWALEncoder(wd)
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

		if err := enc.Encode(msg); err != nil {
			log.Errorf("failed to encode msg: %v", err)
			return
		}
	}
}

func (t *TmWal) ResetNewWalHeight(height int64) {
	writeFile := t.cswal + ".new"
	rd, wd, err := t.open(writeFile)
	if err != nil {
		return
	}
	defer func() {
		t.close(rd)
		t.close(wd)
		util.Rename(t.cswal, writeFile)
		os.RemoveAll(t.mempool)
	}()

	lastHeight := height + 1
	enc := consensus.NewWALEncoder(wd)
	dec := consensus.NewWALDecoder(rd)
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

		if err := enc.Encode(msg); err != nil {
			log.Errorf("failed to encode msg: %v", err)
			return
		}
	}
}

func (t *TmWal) UpdateCsWal(newPath string, c *TmDataStore) {
	writeFile := newPath + "/cs.wal/wal"
	rd, wd, err := t.open(writeFile)
	if err != nil {
		return
	}
	defer func() {
		t.close(rd)
		t.close(wd)
	}()

	dec := cs.NewWALDecoder(rd)
	enc := consensus.NewWALEncoderExt(wd)
	for {
		msg, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("failed to decode msg: %v", err)
			return
		}

		if !cs.FilterBlockWalMessage(c.Height, msg.Msg) {
			continue
		}

		if walEvent := cs.ConvertWalMessage(msg); walEvent != nil {
			log.Infof("event msg: %+v", *msg)
			if err := enc.Encode(walEvent); err != nil {
				log.Errorf("failed to encode msg: %v", err)
				return
			}
		}
	}
}

func (t *TmWal) open(fw string) (rd *os.File, wd *os.File, err error) {
	rd, err = os.Open(t.cswal)
	if err == nil {
		wd, err = os.Create(fw)
		if err != nil {
			log.Errorf("failed to open WAL file: %v", err)
			rd.Close()
		}
	} else {
		log.Errorf("failed to open WAL file: %v", err)
	}
	return
}

func (t *TmWal) close(file *os.File) {
	if file != nil {
		file.Close()
	}
}

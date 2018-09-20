package oldver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/commis/tm-tools/libs/util"
	oldtype "github.com/commis/tm-tools/oldver/types"
	wire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/state"
	dbm "github.com/tendermint/tmlibs/db"
)

func LoadTotalHeight(ldb dbm.DB) int64 {
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
		panic(fmt.Sprintf("Could not unmarshal bytes: %X", bytes))
	}
	return bsj
}

func LoadOldBlock(ldb dbm.DB, height int64) *oldtype.Block {
	meta := LoadOldBlockMeta(ldb, height)

	bytez := []byte{}
	for i := 0; i < meta.BlockID.PartsHeader.Total; i++ {
		part := LoadOldBlockPart(ldb, height, i)
		bytez = append(bytez, part.Bytes...)
	}

	var n int
	var err error
	block := wire.ReadBinary(&oldtype.Block{}, bytes.NewReader(bytez), 0, &n, &err).(*oldtype.Block)
	if err != nil {
		panic(fmt.Sprintf("Error reading block: %v", err))
	}
	return block
}

func LoadOldBlockMeta(ldb dbm.DB, height int64) *oldtype.BlockMeta {
	var n int
	var err error
	r := GetReader(ldb, calcBlockMetaKey(height))
	meta := wire.ReadBinary(&oldtype.BlockMeta{}, r, 0, &n, &err).(*oldtype.BlockMeta)
	if err != nil {
		panic(err)
	}
	return meta
}

func LoadOldBlockPart(ldb dbm.DB, height int64, index int) *oldtype.Part {
	buf := ldb.Get(calcBlockPartKey(height, index))
	if buf == nil {
		return nil
	}

	var n int
	var err error
	part := wire.ReadBinary(&oldtype.Part{}, bytes.NewReader(buf), 0, &n, &err).(*oldtype.Part)
	if err != nil {
		panic(fmt.Sprintf("Error reading block part: %v", err))
	}
	return part
}

func LoadOldBlockCommit(ldb dbm.DB, height int64, prefix string) *oldtype.Commit {
	var buf []byte

	if prefix == "C" {
		buf = ldb.Get(calcBlockCommitKey(height))
	} else if prefix == "SC" {
		buf = ldb.Get(calcSeenCommitKey(height))
	}

	r := bytes.NewReader(buf)
	if r == nil {
		return nil
	}

	var n int
	var err error
	blockCommit := wire.ReadBinary(&oldtype.Commit{}, r, 0, &n, &err).(*oldtype.Commit)
	if err != nil {
		panic(fmt.Sprintf("Error reading commit: %v", err))
	}
	return blockCommit
}

func LoadOldState(ldb dbm.DB) *oldtype.State {
	buf := ldb.Get([]byte(util.StateKey))
	if len(buf) == 0 {
		return nil
	}

	s := &oldtype.State{}
	err := wire.UnmarshalBinary(buf, s)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		cmn.Exit(cmn.Fmt(`LoadState: Data has been corrupted or its spec has changed:%v\n`, err))
	}
	return s
}

func LoadOldAbciResps(ldb dbm.DB) *state.ABCIResponses {
	buf := ldb.Get([]byte(util.ABCIResponsesKey))
	resps := &state.ABCIResponses{}
	err := wire.UnmarshalBinary(buf, resps)
	if err != nil {
		cmn.Exit(cmn.Fmt(`LoadABCIResponses: Data has been corrupted or its spec has changed: %v\n`, err))
	}
	return resps
}

func GetReader(ldb dbm.DB, key []byte) io.Reader {
	bytez := ldb.Get(key)
	if bytez == nil {
		return nil
	}
	return bytes.NewReader(bytez)
}

//==============================================================================
func calcBlockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

func calcBlockCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

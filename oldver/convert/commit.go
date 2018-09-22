package convert

import (
	"bytes"
	"fmt"

	his "github.com/commis/tm-tools/oldver/types"
	oldtype "github.com/commis/tm-tools/oldver/types"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

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
		cmn.Exit(fmt.Sprintf("Error reading commit: %v", err))
	}
	return blockCommit
}

func NewSeenCommit(ldb dbm.DB, height int64, lastBlockID *types.BlockID, nState *state.State) *types.Commit {
	oCommit := LoadOldBlockCommit(ldb, height, "SC")
	return NewCommit(oCommit, lastBlockID, nState)
}

func NewCommit(oCommit *his.Commit, lastBlockID *types.BlockID, nState *state.State) *types.Commit {
	nCommit := &types.Commit{}

	preCommits := []*types.Vote{}
	for i := 0; i < len(oCommit.Precommits); i++ {
		v := oCommit.Precommits[i]
		// node's commit may be nil
		if v == nil {
			preCommits = append(preCommits, nil)
			continue
		}

		one := &types.Vote{
			ValidatorAddress: nState.Validators.Validators[i].Address,
			ValidatorIndex:   i,
			Height:           v.Height,
			Round:            v.Round,
			Timestamp:        v.Timestamp,
			Type:             v.Type,
			BlockID:          *lastBlockID,
			Signature:        v.Signature.Bytes(),
		}
		preCommits = append(preCommits, one)
	}

	nCommit.BlockID = *lastBlockID
	nCommit.Precommits = preCommits

	return nCommit
}

// ====================================================
func calcBlockCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

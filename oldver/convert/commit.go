package convert

import (
	"fmt"

	his "github.com/commis/tm-tools/oldver/types"
	oldtype "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

func LoadOldBlockCommit(ldb dbm.DB, height int64, prefix string) *oldtype.Commit {
	var key []byte
	if prefix == "C" {
		key = calcBlockCommitKey(height)
	} else if prefix == "SC" {
		key = calcSeenCommitKey(height)
	}

	var n int
	var err error
	r := GetReader(ldb, key)
	blockCommit := wire.ReadBinary(&oldtype.Commit{}, r, 0, &n, &err).(*oldtype.Commit)
	if err != nil {
		fmt.Sprintf("Error reading commit: %v", err)
		return nil
	}
	return blockCommit
}

func NewSeenCommit(ldb dbm.DB, height int64, nState *state.State) *types.Commit {
	oCommit := LoadOldBlockCommit(ldb, height, "SC")
	return NewCommit(oCommit, nState)
}

func NewCommit(oCommit *his.Commit, nState *state.State) *types.Commit {
	nCommit := &types.Commit{}
	nCommit.BlockID = NewBlockID(&oCommit.BlockID)

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
			BlockID:          nCommit.BlockID,
			Signature:        v.Signature.Bytes(),
		}
		preCommits = append(preCommits, one)
	}

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

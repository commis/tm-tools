package convert

import (
	"github.com/commis/tm-tools/oldver"
	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

func NewSeenCommit(ldb dbm.DB, height int64, lastBlockID *types.BlockID, nState *state.State) *types.Commit {
	oCommit := oldver.LoadOldBlockCommit(ldb, height, "SC")
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

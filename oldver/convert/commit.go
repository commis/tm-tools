package convert

import (
	"fmt"

	"github.com/commis/tm-tools/libs/log"

	"github.com/tendermint/tendermint/privval"

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

func NewSeenCommit(ldb dbm.DB, height int64, state *state.State, pv *privval.FilePV) *types.Commit {
	oCommit := LoadOldBlockCommit(ldb, height, "SC")
	return NewCommit(oCommit, state, pv)
}

func NewCommit(oCommit *his.Commit, state *state.State, pv *privval.FilePV) *types.Commit {
	nCommit := &types.Commit{}
	nCommit.BlockID = state.LastBlockID

	preCommits := []*types.Vote{}
	for i := 0; i < len(oCommit.Precommits); i++ {
		v := oCommit.Precommits[i]
		// node's commit may be nil
		if v == nil {
			preCommits = append(preCommits, nil)
			continue
		}

		one := &types.Vote{}
		one.ValidatorAddress = state.Validators.Validators[i].Address
		one.ValidatorIndex = i
		one.Height = v.Height
		one.Round = v.Round
		one.Timestamp = v.Timestamp
		one.Type = v.Type
		one.BlockID = nCommit.BlockID

		//重新写签名
		sig, err := pv.PrivKey.Sign(one.SignBytes(state.ChainID))
		if err != nil {
			log.Errorf("failed to sign commit")
			continue
		}
		one.Signature = sig //v.Signature.Bytes()
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

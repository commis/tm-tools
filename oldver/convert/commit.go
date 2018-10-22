package convert

import (
	"fmt"

	"github.com/commis/tm-tools/libs/log"
	"github.com/commis/tm-tools/libs/op/store"
	otp "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

func LoadOldBlockCommit(ldb dbm.DB, height int64, prefix string) *otp.Commit {
	var key []byte
	if prefix == "C" {
		key = calcBlockCommitKey(height)
	} else if prefix == "SC" {
		key = calcSeenCommitKey(height)
	}

	var n int
	var err error
	r := GetReader(ldb, key)
	blockCommit := wire.ReadBinary(&otp.Commit{}, r, 0, &n, &err).(*otp.Commit)
	if err != nil {
		log.Errorf("Error reading commit: %v", err)
		return nil
	}
	return blockCommit
}

func NewSeenCommit(ldb dbm.DB, block *types.Block) *types.Commit {
	oCommit := LoadOldBlockCommit(ldb, block.Height, "SC")
	return NewCommit(oCommit, block)
}

func NewCommit(oCommit *otp.Commit, block *types.Block) *types.Commit {
	nCommit := &types.Commit{}
	nCommit.BlockID = block.LastBlockID

	votes := make([]*types.Vote, 0, len(oCommit.Precommits))
	for _, v := range oCommit.Precommits {
		// node's commit may be nil
		if v == nil {
			votes = append(votes, nil)
			continue
		}

		one := &types.Vote{}
		one.Type = v.Type
		one.BlockID = nCommit.BlockID
		one.Height = v.Height
		one.Round = v.Round
		one.Timestamp = v.Timestamp

		pv := store.GetNodePrivByAddress(v.ValidatorAddress.String())
		sig, _ := pv.PrivVal.PrivKey.Sign(one.SignBytes(block.ChainID))
		one.ValidatorAddress = pv.PrivVal.Address
		one.ValidatorIndex = pv.Index
		one.Signature = sig

		votes = append(votes, one)
	}

	nCommit.Precommits = votes
	return nCommit
}

// ====================================================
func calcBlockCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

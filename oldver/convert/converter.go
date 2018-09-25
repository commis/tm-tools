package convert

import (
	his "github.com/commis/tm-tools/oldver/types"
	gco "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

func CvtNewPubKey(old gco.PubKey) crypto.PubKey {
	oBytes := old.Unwrap().(gco.PubKeyEd25519)
	nBytes := ed25519.PubKeyEd25519{}
	for i, bt := range oBytes {
		nBytes[i] = bt
	}
	//fmt.Println(oBytes.String())
	//fmt.Println(nBytes.String())
	return nBytes
}

func CvtNewPrivKey(old gco.PrivKey) crypto.PrivKey {
	oBytes := old.Unwrap().(gco.PrivKeyEd25519)
	nBytes := ed25519.PrivKeyEd25519{}
	for i, bt := range oBytes {
		nBytes[i] = bt
	}
	return nBytes
}

func CvtNewSignature(old gco.Signature) []byte {
	oBytes := old.Unwrap().(gco.SignatureEd25519)
	nBytes := make([]byte, len(oBytes))
	for i, bt := range oBytes {
		nBytes[i] = bt
	}
	return nBytes
}

func CvtNewVote(old *his.Vote) *types.Vote {
	vote := new(types.Vote)

	vote.ValidatorAddress = old.ValidatorAddress.Bytes()
	vote.ValidatorIndex = old.ValidatorIndex
	vote.Height = old.Height
	vote.Round = old.Round
	vote.Timestamp = old.Timestamp
	vote.Type = old.Type
	vote.BlockID = types.BlockID{
		Hash: old.BlockID.Hash.Bytes(),
		PartsHeader: types.PartSetHeader{
			Total: old.BlockID.PartsHeader.Total,
			Hash:  old.BlockID.PartsHeader.Hash.Bytes(),
		}}
	vote.Signature = CvtNewSignature(old.Signature)

	return vote
}

func CvtNewEvidence(old his.Evidence) types.Evidence {
	oEvi := old.(*his.DuplicateVoteEvidence)

	evidence := &types.DuplicateVoteEvidence{}
	evidence.PubKey = CvtNewPubKey(oEvi.PubKey)
	evidence.VoteA = CvtNewVote(oEvi.VoteA)
	evidence.VoteB = CvtNewVote(oEvi.VoteB)

	return evidence
}

func CvtValidatorsInfo(old *his.ValidatorsInfo) *state.ValidatorsInfo {
	valInfo := new(state.ValidatorsInfo)
	valInfo.ValidatorSet = NewValidatorSet(old.ValidatorSet)
	valInfo.LastHeightChanged = old.LastHeightChanged

	return valInfo
}

func CvtConsensusParamsInfo(old *his.ConsensusParamsInfo) *state.ConsensusParamsInfo {
	consensusParam := new(state.ConsensusParamsInfo)
	consensusParam.ConsensusParams = CvtConsensusParams(&old.ConsensusParams)
	consensusParam.LastHeightChanged = old.LastHeightChanged

	return consensusParam
}

func CvtConsensusParams(old *his.ConsensusParams) types.ConsensusParams {
	return types.ConsensusParams{
		BlockSize:      types.BlockSize{MaxBytes: old.BlockSize.MaxBytes, MaxTxs: old.BlockSize.MaxTxs, MaxGas: old.BlockSize.MaxGas},
		TxSize:         types.TxSize{MaxBytes: old.TxSize.MaxBytes, MaxGas: old.TxSize.MaxGas},
		BlockGossip:    types.BlockGossip{BlockPartSizeBytes: old.BlockGossip.BlockPartSizeBytes},
		EvidenceParams: types.EvidenceParams{MaxAge: old.EvidenceParams.MaxAge},
	}
}

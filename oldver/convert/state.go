package convert

import (
	"github.com/commis/tm-tools/libs/util"
	oldver "github.com/commis/tm-tools/oldver"
	his "github.com/commis/tm-tools/oldver/types"
	gco "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

func InitState(ldb dbm.DB) *state.State {
	oState := oldver.LoadOldState(ldb)
	retState := &state.State{}
	retState.ChainID = oState.ChainID
	retState.LastBlockHeight = oState.LastBlockHeight
	retState.LastBlockTotalTx = oState.LastBlockTotalTx
	retState.LastBlockID = NewBlockID(&oState.LastBlockID)
	retState.LastBlockTime = oState.LastBlockTime
	retState.Validators = NewValidatorSet(oState.Validators)
	retState.LastValidators = NewValidatorSet(oState.LastValidators)
	retState.LastHeightValidatorsChanged = oState.LastHeightValidatorsChanged
	retState.ConsensusParams = NewConsensusParams(&oState.ConsensusParams)
	retState.LastHeightConsensusParamsChanged = oState.LastHeightConsensusParamsChanged
	retState.LastResultsHash = oState.LastResultsHash
	retState.AppHash = oState.AppHash

	return retState
}

func SaveState(ldb db.DB, lastBlockID *types.BlockID, nState *state.State) {
	nState.LastBlockID = *lastBlockID
	util.SaveNewState(ldb, nState)
}

func NewBlockID(old *his.BlockID) types.BlockID {
	return types.BlockID{
		Hash:        old.Hash.Bytes(),
		PartsHeader: types.PartSetHeader{Total: old.PartsHeader.Total, Hash: old.PartsHeader.Hash.Bytes()},
	}
}

func NewConsensusParams(old *his.ConsensusParams) types.ConsensusParams {
	return types.ConsensusParams{
		BlockSize:      types.BlockSize{MaxBytes: old.BlockSize.MaxBytes, MaxTxs: old.BlockSize.MaxTxs, MaxGas: old.BlockSize.MaxGas},
		TxSize:         types.TxSize{MaxBytes: old.TxSize.MaxBytes, MaxGas: old.TxSize.MaxGas},
		BlockGossip:    types.BlockGossip{BlockPartSizeBytes: old.BlockGossip.BlockPartSizeBytes},
		EvidenceParams: types.EvidenceParams{MaxAge: old.EvidenceParams.MaxAge},
	}
}

func NewGenesisDoc(old *his.GenesisDoc) *types.GenesisDoc {
	newGenesisDoc := &types.GenesisDoc{
		AppHash:     old.AppHash.Bytes(),
		ChainID:     old.ChainID,
		GenesisTime: old.GenesisTime,
		Validators:  []types.GenesisValidator{},
	}
	for _, val := range old.Validators {
		one := types.GenesisValidator{}
		one.Power = val.Power
		one.Name = val.Name
		one.PubKey = NewPubKey(val.PubKey)

		newGenesisDoc.Validators = append(newGenesisDoc.Validators, one)
	}

	return newGenesisDoc
}

func NewValidatorSet(oValidatorSet *his.ValidatorSet) *types.ValidatorSet {
	nValidatorSet := &types.ValidatorSet{
		Validators: []*types.Validator{},
	}

	// Validators
	for _, val := range oValidatorSet.Validators {
		one := &types.Validator{}
		one.Accum = val.Accum
		one.Address = val.Address.Bytes()
		one.PubKey = NewPubKey(val.PubKey)
		one.VotingPower = val.VotingPower

		nValidatorSet.Validators = append(nValidatorSet.Validators, one)
	}

	// Proposer
	nValidatorSet.Proposer = &types.Validator{
		Address:     oValidatorSet.Proposer.Address.Bytes(),
		PubKey:      NewPubKey(oValidatorSet.Proposer.PubKey),
		VotingPower: oValidatorSet.Proposer.VotingPower,
		Accum:       oValidatorSet.Proposer.Accum,
	}

	return nValidatorSet
}

func NewPubKey(old gco.PubKey) crypto.PubKey {
	oBytes := old.Unwrap().(gco.PubKeyEd25519)
	nBytes := ed25519.PubKeyEd25519{}
	for i, bt := range oBytes {
		nBytes[i] = bt
	}
	//fmt.Println(oBytes.String())
	//fmt.Println(nBytes.String())
	return nBytes
}

func NewPrivKey(old gco.PrivKey) crypto.PrivKey {
	oBytes := old.Unwrap().(gco.PrivKeyEd25519)
	nBytes := ed25519.PrivKeyEd25519{}
	for i, bt := range oBytes {
		nBytes[i] = bt
	}
	return nBytes
}

func NewSignature(old gco.Signature) []byte {
	oBytes := old.Unwrap().(gco.SignatureEd25519)
	nBytes := make([]byte, len(oBytes))
	for i, bt := range oBytes {
		nBytes[i] = bt
	}
	return nBytes
}

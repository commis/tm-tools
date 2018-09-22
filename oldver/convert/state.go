package convert

import (
	"encoding/json"
	"log"

	"github.com/commis/tm-tools/libs/util"
	his "github.com/commis/tm-tools/oldver/types"
	oldtype "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func SaveOldBlockStoreStateJson(db dbm.DB, bsj his.BlockStoreStateJSON) {
	bytes, err := json.Marshal(bsj)
	if err != nil {
		cmn.PanicSanity(cmn.Fmt("Could not marshal state bytes: %v", err))
	}
	db.SetSync([]byte(util.BlockStoreKey), bytes)
}

func InitState(ldb dbm.DB) *state.State {
	oState := LoadOldState(ldb)
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

func ConvertValidatorsInfo(old *his.ValidatorsInfo) *state.ValidatorsInfo {
	valInfo := new(state.ValidatorsInfo)
	valInfo.ValidatorSet = NewValidatorSet(old.ValidatorSet)
	valInfo.LastHeightChanged = old.LastHeightChanged

	return valInfo
}

func NewValidatorSet(oValidatorSet *his.ValidatorSet) *types.ValidatorSet {
	if oValidatorSet == nil {
		return nil
	}

	nValidatorSet := &types.ValidatorSet{
		Validators: []*types.Validator{},
	}

	// Validators
	for _, val := range oValidatorSet.Validators {
		one := &types.Validator{}
		one.Accum = val.Accum
		one.Address = val.Address.Bytes()
		one.PubKey = CvtNewPubKey(val.PubKey)
		one.VotingPower = val.VotingPower

		nValidatorSet.Validators = append(nValidatorSet.Validators, one)
	}

	// Proposer
	nValidatorSet.Proposer = &types.Validator{
		Address:     oValidatorSet.Proposer.Address.Bytes(),
		PubKey:      CvtNewPubKey(oValidatorSet.Proposer.PubKey),
		VotingPower: oValidatorSet.Proposer.VotingPower,
		Accum:       oValidatorSet.Proposer.Accum,
	}

	return nValidatorSet
}

func LoadOldABCIResponse(db dbm.DB, height int64) *his.ABCIResponses {
	buf := db.Get(util.CalcABCIResponsesKey(height))
	if len(buf) == 0 {
		return nil
	}

	abciResponses := new(his.ABCIResponses)
	err := wire.UnmarshalBinary(buf, abciResponses)
	if err != nil {
		//fmt.Printf("LoadABCIResponses: Data has been corrupted or its spec has changed: %v\n", err)
		return nil
	}

	return abciResponses
}

func SaveNewABCIResponse(ldb dbm.DB, ndb db.DB, height int64) {
	abciResponses := LoadOldABCIResponse(ldb, height)
	if abciResponses != nil {
		ndb.SetSync(util.CalcABCIResponsesKey(height), abciResponses.Bytes())
	}
}

func DeleteABCIResponse(newVer bool, ldb dbm.DB, ndb db.DB, height int64) {
	key := util.CalcABCIResponsesKey(height)
	if newVer {
		if ndb.Has(key) {
			ndb.DeleteSync(key)
		}
	} else {
		if ldb.Has(key) {
			ldb.DeleteSync(key)
		}
	}
}

func LoadOldConsensusParamsInfo(db dbm.DB, height int64) *his.ConsensusParamsInfo {
	buf := db.Get(util.CalcConsensusParamsKey(height))
	if len(buf) == 0 {
		return nil
	}

	paramsInfo := new(his.ConsensusParamsInfo)
	err := wire.UnmarshalBinary(buf, paramsInfo)
	if err != nil {
		return nil
	}

	return paramsInfo
}

func SaveNewConsensusParams(ldb dbm.DB, ndb db.DB, height int64) {
	paramsInfo := LoadOldConsensusParamsInfo(ldb, height)
	if paramsInfo != nil {
		ndb.SetSync(util.CalcConsensusParamsKey(height), paramsInfo.Bytes())
	}
}

func DeleteConsensusParam(newVer bool, ldb dbm.DB, ndb db.DB, height int64) {
	key := util.CalcConsensusParamsKey(height)
	if newVer {
		if ndb.Has(key) {
			ndb.DeleteSync(key)
		}
	} else {
		if ldb.Has(key) {
			ldb.DeleteSync(key)
		}
	}
}

func LoadOldValidatorsInfo(db dbm.DB, height int64) *his.ValidatorsInfo {
	buf := db.Get(util.CalcValidatorsKey(height))
	if len(buf) == 0 {
		return nil
	}

	v := new(his.ValidatorsInfo)
	err := wire.UnmarshalBinary(buf, v)
	if err != nil {
		log.Printf("LoadValidators: Data has been corrupted or its spec has changed: %v\n", err)
		return nil
	}

	return v
}

func SaveNewValidators(ldb dbm.DB, ndb db.DB, height int64) {
	valInfo := LoadOldValidatorsInfo(ldb, height)
	if valInfo != nil {
		nValInfo := ConvertValidatorsInfo(valInfo)
		ndb.SetSync(util.CalcValidatorsKey(height), nValInfo.Bytes())
	}
}

func DeleteValidator(newVer bool, ldb dbm.DB, ndb db.DB, height int64) {
	key := util.CalcValidatorsKey(height)
	if newVer {
		if ndb.Has(key) {
			ndb.DeleteSync(key)
		}
	} else {
		if ldb.Has(key) {
			ldb.DeleteSync(key)
		}
	}
}

package oldver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	his "github.com/commis/tm-tools/oldver/types"
	wire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	GenesisDocKey = "genesisDoc"
)

func LoadOldABCIResponses(db dbm.DB, height int64) *his.ABCIResponses {
	buf := db.Get(calcABCIResponsesKey(height))
	if len(buf) == 0 {
		return nil
	}
	fmt.Println(string(buf))

	abciResponses := new(his.ABCIResponses)
	err := wire.UnmarshalBinary(buf, abciResponses)
	if err != nil {
		log.Printf("LoadABCIResponses: Data has been corrupted or its spec has changed: %v\n", err)
		return nil
	}

	return abciResponses
}

func SaveNewABCIResponses(ldb dbm.DB, ndb db.DB, height int64) {
	abciResponses := LoadOldABCIResponses(ldb, height)
	if abciResponses != nil {
		ndb.SetSync(calcABCIResponsesKey(height), abciResponses.Bytes())
	}
}

func LoadOldConsensusParamsInfo(db dbm.DB, height int64) *his.ConsensusParamsInfo {
	buf := db.Get(calcConsensusParamsKey(height))
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
		ndb.SetSync(calcConsensusParamsKey(height), paramsInfo.Bytes())
	}
}

func LoadOldValidatorsInfo(db dbm.DB, height int64) *his.ValidatorsInfo {
	buf := db.Get(calcValidatorsKey(height))
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
		ndb.SetSync(calcValidatorsKey(height), valInfo.Bytes())
	}
}

func LoadOldGenesisDoc(db dbm.DB) (*his.GenesisDoc, error) {
	bytes := db.Get([]byte(GenesisDocKey))
	if len(bytes) == 0 {
		return nil, errors.New("genesis doc not found")
	}

	var genDoc *his.GenesisDoc
	err := json.Unmarshal(bytes, &genDoc)
	if err != nil {
		log.Printf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, bytes)
	}
	return genDoc, nil
}

func SaveNewGenesisDoc(ldb db.DB, genDoc *types.GenesisDoc) {
	bytes, err := json.Marshal(genDoc)
	if err != nil {
		log.Printf("Failed to save genesis doc due to marshaling error: %v", err)
		return
	}
	ldb.SetSync([]byte(GenesisDocKey), bytes)
}

//------------------------------------------------------------------------
func calcValidatorsKey(height int64) []byte {
	return []byte(cmn.Fmt("validatorsKey:%v", height))
}

func calcConsensusParamsKey(height int64) []byte {
	return []byte(cmn.Fmt("consensusParamsKey:%v", height))
}

func calcABCIResponsesKey(height int64) []byte {
	return []byte(cmn.Fmt("abciResponsesKey:%v", height))
}

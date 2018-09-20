package types

import (
	"encoding/json"
	"time"

	gco "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

type GenesisDoc struct {
	GenesisTime     time.Time          `json:"genesis_time"`
	ChainID         string             `json:"chain_id"`
	ConsensusParams *ConsensusParams   `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator `json:"validators"`
	AppHash         cmn.HexBytes       `json:"app_hash"`
	AppStateJSON    json.RawMessage    `json:"app_state,omitempty"`
	AppOptions      json.RawMessage    `json:"app_options,omitempty"` // DEPRECATED
}

type GenesisValidator struct {
	PubKey gco.PubKey `json:"pub_key"`
	Power  int64      `json:"power"`
	Name   string     `json:"name"`
}

func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	genDoc := GenesisDoc{}
	err := json.Unmarshal(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}
	return &genDoc, err
}

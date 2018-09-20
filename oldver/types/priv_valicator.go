package types

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	gco "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

type PrivValidatorFS struct {
	Address       Address       `json:"address"`
	PubKey        gco.PubKey    `json:"pub_key"`
	LastHeight    int64         `json:"last_height"`
	LastRound     int           `json:"last_round"`
	LastStep      int8          `json:"last_step"`
	LastSignature gco.Signature `json:"last_signature,omitempty"` // so we dont lose signatures
	LastSignBytes cmn.HexBytes  `json:"last_signbytes,omitempty"` // so we dont lose signatures
	PrivKey       gco.PrivKey   `json:"priv_key"`

	//Signer   `json:"-"`
	filePath string
	mtx      sync.Mutex
}

func LoadPrivValidator(filePath string) *PrivValidatorFS {
	privValJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		cmn.Exit(err.Error())
	}
	privVal := &PrivValidatorFS{}
	err = json.Unmarshal(privValJSONBytes, &privVal)
	if err != nil {
		cmn.Exit(cmn.Fmt("Error reading PrivValidator from %v: %v\n", filePath, err))
	}

	privVal.filePath = filePath
	return privVal
}

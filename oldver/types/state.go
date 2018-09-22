package types

import (
	"time"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
)

type State struct {
	ChainID                          string
	LastBlockHeight                  int64
	LastBlockTotalTx                 int64
	LastBlockID                      BlockID
	LastBlockTime                    time.Time
	Validators                       *ValidatorSet
	LastValidators                   *ValidatorSet
	LastHeightValidatorsChanged      int64
	ConsensusParams                  ConsensusParams
	LastHeightConsensusParamsChanged int64
	LastResultsHash                  []byte
	AppHash                          []byte
}

func (s State) Bytes() []byte {
	bz, err := wire.MarshalBinary(s)
	if err != nil {
		panic(err)
	}
	return bz
}

// ABCIResponse
type ABCIResponses struct {
	DeliverTx []*abci.ResponseDeliverTx
	EndBlock  *abci.ResponseEndBlock
}

func (a *ABCIResponses) Bytes() []byte {
	bz, err := wire.MarshalBinary(*a)
	if err != nil {
		panic(err)
	}
	return bz
}

// ConsensusParamsInfo
type ConsensusParamsInfo struct {
	ConsensusParams   ConsensusParams
	LastHeightChanged int64
}

func (params ConsensusParamsInfo) Bytes() []byte {
	bz, err := wire.MarshalBinary(params)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidatorsInfo
type ValidatorsInfo struct {
	ValidatorSet      *ValidatorSet
	LastHeightChanged int64
}

func (valInfo *ValidatorsInfo) Bytes() []byte {
	bz, err := wire.MarshalBinary(*valInfo)
	if err != nil {
		panic(err)
	}
	return bz
}

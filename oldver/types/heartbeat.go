package types

import crypto "github.com/tendermint/go-crypto"

type Heartbeat struct {
	ValidatorAddress Address          `json:"validator_address"`
	ValidatorIndex   int              `json:"validator_index"`
	Height           int64            `json:"height"`
	Round            int              `json:"round"`
	Sequence         int              `json:"sequence"`
	Signature        crypto.Signature `json:"signature"`
}

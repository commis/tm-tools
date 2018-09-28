package types

import (
	"time"

	crypto "github.com/tendermint/go-crypto"
)

type Proposal struct {
	Height           int64            `json:"height"`
	Round            int              `json:"round"`
	Timestamp        time.Time        `json:"timestamp"`
	BlockPartsHeader PartSetHeader    `json:"block_parts_header"`
	POLRound         int              `json:"pol_round"`    // -1 if null.
	POLBlockID       BlockID          `json:"pol_block_id"` // zero if null.
	Signature        crypto.Signature `json:"signature"`
}

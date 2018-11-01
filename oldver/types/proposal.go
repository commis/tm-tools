package types

import (
	"time"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
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

func (p *Proposal) SignBytes(chainID string) []byte {
	bz, err := wire.MarshalJSON(CanonicalJSONOnceProposal{
		ChainID:  chainID,
		Proposal: CanonicalProposal(p),
	})
	if err != nil {
		panic(err)
	}
	return bz
}

type CanonicalJSONOnceProposal struct {
	ChainID  string                `json:"chain_id"`
	Proposal CanonicalJSONProposal `json:"proposal"`
}

type CanonicalJSONProposal struct {
	BlockPartsHeader CanonicalJSONPartSetHeader `json:"block_parts_header"`
	Height           int64                      `json:"height"`
	POLBlockID       CanonicalJSONBlockID       `json:"pol_block_id"`
	POLRound         int                        `json:"pol_round"`
	Round            int                        `json:"round"`
	Timestamp        string                     `json:"timestamp"`
}

func CanonicalProposal(proposal *Proposal) CanonicalJSONProposal {
	return CanonicalJSONProposal{
		BlockPartsHeader: CanonicalPartSetHeader(proposal.BlockPartsHeader),
		Height:           proposal.Height,
		Timestamp:        CanonicalTime(proposal.Timestamp),
		POLBlockID:       CanonicalBlockID(proposal.POLBlockID),
		POLRound:         proposal.POLRound,
		Round:            proposal.Round,
	}
}

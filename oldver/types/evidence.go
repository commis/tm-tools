package types

import (
	cmn "github.com/tendermint/tmlibs/common"
)

type Evidence interface {
	Height() int64               // height of the equivocation
	Address() []byte             // address of the equivocating validator
	Index() int                  // index of the validator in the validator set
	Hash() []byte                // hash of the evidence
	Verify(chainID string) error // verify the evidence
	Equal(Evidence) bool         // check equality of evidence

	String() string
}

type EvidenceList []Evidence

type EvidenceData struct {
	Evidence EvidenceList `json:"evidence"`

	// Volatile
	hash cmn.HexBytes
}

type EvidenceInfo struct {
	Committed bool
	Priority  int64
	Evidence  Evidence
}

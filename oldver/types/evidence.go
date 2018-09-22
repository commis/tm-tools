package types

import (
	"bytes"
	"errors"
	"fmt"

	gco "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
	"golang.org/x/crypto/ripemd160"
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

type DuplicateVoteEvidence struct {
	PubKey gco.PubKey
	VoteA  *Vote
	VoteB  *Vote
}

func (dve *DuplicateVoteEvidence) String() string {
	return fmt.Sprintf("VoteA: %v; VoteB: %v", dve.VoteA, dve.VoteB)
}

// Height returns the height this evidence refers to.
func (dve *DuplicateVoteEvidence) Height() int64 {
	return dve.VoteA.Height
}

// Address returns the address of the validator.
func (dve *DuplicateVoteEvidence) Address() []byte {
	return dve.PubKey.Address()
}

// Index returns the index of the validator.
func (dve *DuplicateVoteEvidence) Index() int {
	return dve.VoteA.ValidatorIndex
}

// Hash returns the hash of the evidence.
func (dve *DuplicateVoteEvidence) Hash() []byte {
	return wireHasher(dve).Hash()
}

// Verify returns an error if the two votes aren't conflicting.
// To be conflicting, they must be from the same validator, for the same H/R/S, but for different blocks.
func (dve *DuplicateVoteEvidence) Verify(chainID string) error {
	// H/R/S must be the same
	if dve.VoteA.Height != dve.VoteB.Height ||
		dve.VoteA.Round != dve.VoteB.Round ||
		dve.VoteA.Type != dve.VoteB.Type {
		return fmt.Errorf("DuplicateVoteEvidence Error: H/R/S does not match. Got %v and %v", dve.VoteA, dve.VoteB)
	}

	// Address must be the same
	if !bytes.Equal(dve.VoteA.ValidatorAddress, dve.VoteB.ValidatorAddress) {
		return fmt.Errorf("DuplicateVoteEvidence Error: Validator addresses do not match. Got %X and %X", dve.VoteA.ValidatorAddress, dve.VoteB.ValidatorAddress)
	}
	// XXX: Should we enforce index is the same ?
	if dve.VoteA.ValidatorIndex != dve.VoteB.ValidatorIndex {
		return fmt.Errorf("DuplicateVoteEvidence Error: Validator indices do not match. Got %d and %d", dve.VoteA.ValidatorIndex, dve.VoteB.ValidatorIndex)
	}

	// BlockIDs must be different
	if dve.VoteA.BlockID.Equals(dve.VoteB.BlockID) {
		return fmt.Errorf("DuplicateVoteEvidence Error: BlockIDs are the same (%v) - not a real duplicate vote", dve.VoteA.BlockID)
	}

	// Signatures must be valid
	if !dve.PubKey.VerifyBytes(dve.VoteA.SignBytes(chainID), dve.VoteA.Signature) {
		return fmt.Errorf("DuplicateVoteEvidence Error verifying VoteA: %v", errors.New("Invalid signature"))
	}
	if !dve.PubKey.VerifyBytes(dve.VoteB.SignBytes(chainID), dve.VoteB.Signature) {
		return fmt.Errorf("DuplicateVoteEvidence Error verifying VoteB: %v", errors.New("Invalid signature"))
	}

	return nil
}

// Equal checks if two pieces of evidence are equal.
func (dve *DuplicateVoteEvidence) Equal(ev Evidence) bool {
	if _, ok := ev.(*DuplicateVoteEvidence); !ok {
		return false
	}

	// just check their hashes
	dveHash := wireHasher(dve).Hash()
	evHash := wireHasher(ev).Hash()
	return bytes.Equal(dveHash, evHash)
}

type hasher struct {
	item interface{}
}

func (h hasher) Hash() []byte {
	hasher := ripemd160.New()
	bz, err := wire.MarshalBinary(h.item)
	if err != nil {
		panic(err)
	}
	_, err = hasher.Write(bz)
	if err != nil {
		panic(err)
	}
	return hasher.Sum(nil)

}

func tmHash(item interface{}) []byte {
	h := hasher{item}
	return h.Hash()
}

func wireHasher(item interface{}) merkle.Hasher {
	return hasher{item}
}

package types

import (
	"sync"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
	"golang.org/x/crypto/ripemd160"
)

type Part struct {
	Index int                `json:"index"`
	Bytes cmn.HexBytes       `json:"bytes"`
	Proof merkle.SimpleProof `json:"proof"`
	// Cache
	hash []byte
}

func (part *Part) Hash() []byte {
	if part.hash != nil {
		return part.hash
	}
	hasher := ripemd160.New()
	hasher.Write(part.Bytes) // nolint: errcheck, gas
	part.hash = hasher.Sum(nil)
	return part.hash
}

type PartSet struct {
	total int
	hash  []byte

	mtx           sync.Mutex
	parts         []*Part
	partsBitArray *cmn.BitArray
	count         int
}

func (ps *PartSet) Total() int {
	if ps == nil {
		return 0
	}
	return ps.total
}

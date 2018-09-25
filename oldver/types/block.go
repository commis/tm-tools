package types

import (
	"bytes"
	"time"

	gco "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type BlockID struct {
	Hash        cmn.HexBytes  `json:"hash"`
	PartsHeader PartSetHeader `json:"parts"`
}

func (blockID BlockID) Equals(other BlockID) bool {
	return bytes.Equal(blockID.Hash, other.Hash) &&
		blockID.PartsHeader.Equals(other.PartsHeader)
}

type PartSetHeader struct {
	Total int          `json:"total"`
	Hash  cmn.HexBytes `json:"hash"`
}

func (psh PartSetHeader) Equals(other PartSetHeader) bool {
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash)
}

type Data struct {
	Txs  Txs `json:"txs"`
	hash cmn.HexBytes
}

type Tx []byte
type Txs []Tx

type Commit struct {
	BlockID    BlockID `json:"blockID"`
	Precommits []*Vote `json:"precommits"`

	// Volatile
	firstPrecommit *Vote
	hash           cmn.HexBytes
	bitArray       *cmn.BitArray
}

type Address = cmn.HexBytes

type Vote struct {
	ValidatorAddress Address       `json:"validator_address"`
	ValidatorIndex   int           `json:"validator_index"`
	Height           int64         `json:"height"`
	Round            int           `json:"round"`
	Timestamp        time.Time     `json:"timestamp"`
	Type             byte          `json:"type"`
	BlockID          BlockID       `json:"block_id"` // zero if vote is nil.
	Signature        gco.Signature `json:"signature"`
}

func (vote *Vote) SignBytes(chainID string) []byte {
	bz, err := wire.MarshalJSON(CanonicalJSONOnceVote{
		chainID,
		CanonicalVote(vote),
	})
	if err != nil {
		panic(err)
	}
	return bz
}

type CanonicalJSONOnceVote struct {
	ChainID string            `json:"chain_id"`
	Vote    CanonicalJSONVote `json:"vote"`
}

type CanonicalJSONVote struct {
	BlockID   CanonicalJSONBlockID `json:"block_id"`
	Height    int64                `json:"height"`
	Round     int                  `json:"round"`
	Timestamp string               `json:"timestamp"`
	Type      byte                 `json:"type"`
}

type CanonicalJSONBlockID struct {
	Hash        cmn.HexBytes               `json:"hash,omitempty"`
	PartsHeader CanonicalJSONPartSetHeader `json:"parts,omitempty"`
}

func CanonicalVote(vote *Vote) CanonicalJSONVote {
	return CanonicalJSONVote{
		BlockID:   CanonicalBlockID(vote.BlockID),
		Height:    vote.Height,
		Round:     vote.Round,
		Timestamp: CanonicalTime(vote.Timestamp),
		Type:      vote.Type,
	}
}

const TimeFormat = wire.RFC3339Millis

func CanonicalTime(t time.Time) string {
	return t.UTC().Format(TimeFormat)
}

func CanonicalBlockID(blockID BlockID) CanonicalJSONBlockID {
	return CanonicalJSONBlockID{
		Hash:        blockID.Hash,
		PartsHeader: CanonicalPartSetHeader(blockID.PartsHeader),
	}
}

func CanonicalPartSetHeader(psh PartSetHeader) CanonicalJSONPartSetHeader {
	return CanonicalJSONPartSetHeader{
		psh.Hash,
		psh.Total,
	}
}

type CanonicalJSONPartSetHeader struct {
	Hash  cmn.HexBytes `json:"hash"`
	Total int          `json:"total"`
}

type Block struct {
	*Header    `json:"header"`
	*Data      `json:"data"`
	Evidence   EvidenceData `json:"evidence"`
	LastCommit *Commit      `json:"last_commit"`
}

func (b *Block) MakePartSet(partSize int) *PartSet {
	bz, err := wire.MarshalBinary(b)
	if err != nil {
		panic(err)
	}
	return NewPartSetFromData(bz, partSize)
}

func NewPartSetFromData(data []byte, partSize int) *PartSet {
	// divide data into 4kb parts.
	total := (len(data) + partSize - 1) / partSize
	parts := make([]*Part, total)
	parts_ := make([]merkle.Hasher, total)
	partsBitArray := cmn.NewBitArray(total)
	for i := 0; i < total; i++ {
		part := &Part{
			Index: i,
			Bytes: data[i*partSize : cmn.MinInt(len(data), (i+1)*partSize)],
		}
		parts[i] = part
		parts_[i] = part
		partsBitArray.SetIndex(i, true)
	}
	// Compute merkle proofs
	root, proofs := merkle.SimpleProofsFromHashers(parts_)
	for i := 0; i < total; i++ {
		parts[i].Proof = *proofs[i]
	}
	return &PartSet{
		total:         total,
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         total,
	}
}

type BlockStoreStateJSON struct {
	Height int64
}

type BlockMeta struct {
	BlockID BlockID `json:"block_id"` // the block hash and partsethash
	Header  *Header `json:"header"`   // The block's Header
}

type Header struct {
	// basic block info
	ChainID string    `json:"chain_id"`
	Height  int64     `json:"height"`
	Time    time.Time `json:"time"`
	NumTxs  int64     `json:"num_txs"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`
	TotalTxs    int64   `json:"total_txs"`

	// hashes of block data
	LastCommitHash cmn.HexBytes `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       cmn.HexBytes `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash  cmn.HexBytes `json:"validators_hash"`   // validators for the current block
	ConsensusHash   cmn.HexBytes `json:"consensus_hash"`    // consensus params for current block
	AppHash         cmn.HexBytes `json:"app_hash"`          // state after txs from the previous block
	LastResultsHash cmn.HexBytes `json:"last_results_hash"` // root hash of all results from the txs from the previous block

	// consensus info
	EvidenceHash cmn.HexBytes `json:"evidence_hash"` // evidence included in the block
}

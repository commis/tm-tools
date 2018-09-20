package types

import (
	"time"

	gco "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

type BlockID struct {
	Hash        cmn.HexBytes  `json:"hash"`
	PartsHeader PartSetHeader `json:"parts"`
}

type PartSetHeader struct {
	Total int          `json:"total"`
	Hash  cmn.HexBytes `json:"hash"`
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

type Block struct {
	*Header    `json:"header"`
	*Data      `json:"data"`
	Evidence   EvidenceData `json:"evidence"`
	LastCommit *Commit      `json:"last_commit"`
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

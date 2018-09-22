package convert

import (
	"fmt"
	"log"

	"github.com/commis/tm-tools/libs/util"
	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/evidence"
	"github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tmlibs/db"
)

const (
	baseKeyLookup   = "evidence-lookup"   // all evidence
	baseKeyOutqueue = "evidence-outqueue" // not-yet broadcast
	baseKeyPending  = "evidence-pending"  // broadcast but not committed
)

func UpgradeEvidence(ldb dbm.DB, ndb db.DB) {
	NewEvidence(ldb, ndb, baseKeyLookup)
	NewEvidence(ldb, ndb, baseKeyOutqueue)
	NewEvidence(ldb, ndb, baseKeyPending)
}

func NewEvidence(ldb dbm.DB, ndb db.DB, prefixKey string) {
	cnt := 0
	limit := 1000
	batch := ndb.NewBatch()
	iter := dbm.IteratePrefix(ldb, []byte(prefixKey))
	for ; iter.Valid(); iter.Next() {
		var ei his.EvidenceInfo
		wire.ReadBinaryBytes(iter.Value(), &ei)
		key := GetEvidenceKey(prefixKey, ei.Evidence.Height(), ei.Evidence.Hash())

		newEi := evidence.EvidenceInfo{
			Committed: ei.Committed,
			Priority:  ei.Priority,
			Evidence:  CvtNewEvidence(ei.Evidence)}

		util.SaveNewEvidence(batch, key, &newEi)
		if cnt%limit == 0 {
			log.Printf("batch write evidence %v\n", cnt)
			batch.Write()
			batch = ndb.NewBatch()
		}
	}
	if cnt%limit != 0 {
		log.Printf("batch write evidence %v\n", cnt)
		batch.Write()
	}
}

func GetEvidenceKey(prefixKey string, height int64, hash []byte) []byte {
	return _key("%s/%s/%X", prefixKey, bE(height), hash)
}

// big endian padded hex
func bE(h int64) string {
	return fmt.Sprintf("%0.16X", h)
}

func _key(fmt_ string, o ...interface{}) []byte {
	return []byte(fmt.Sprintf(fmt_, o...))
}

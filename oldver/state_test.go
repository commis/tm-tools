package oldver_test

import (
	"encoding/json"
	"testing"

	"github.com/commis/tm-tools/oldver"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	stateDbm dbm.DB
)

func onTestStart() {
	stateDbm = dbm.NewDB("state", dbm.LevelDBBackend, "/home/share/tm-v0.18.0/tendermint/data")
}

func onTestStop() {
	stateDbm.Close()
}

func TestLoadState(t *testing.T) {
	onTestStart()
	defer onTestStop()

	level := stateDbm.(*dbm.GoLevelDB).DB()
	query := level.NewIterator(nil, nil)
	defer query.Release()

	query.Seek([]byte(""))
	for query.Next() {
		t.Log(string(query.Key()))
	}
}

func TestLoadABCIResponses(t *testing.T) {
	onTestStart()
	defer onTestStop()

	var startHeight int64 = 1
	var totalHeight int64 = 47
	for i := startHeight; i <= totalHeight; i++ {
		abciResp := oldver.LoadOldABCIResponses(stateDbm, i)
		if abciResp != nil {
			res, _ := json.Marshal(abciResp)
			t.Log(string(res))
		}
	}
}

func TestLoadOldConsensusParamsInfo(t *testing.T) {
	onTestStart()
	defer onTestStop()

	var startHeight int64 = 1
	var totalHeight int64 = 47
	for i := startHeight; i <= totalHeight; i++ {
		consensusParamsInfo := oldver.LoadOldConsensusParamsInfo(stateDbm, i)
		if consensusParamsInfo != nil {
			res, _ := json.Marshal(consensusParamsInfo)
			t.Log(string(res))
		}
	}
}

func TestLoadOldValidatorsInfo(t *testing.T) {
	onTestStart()
	defer onTestStop()

	var startHeight int64 = 1
	var totalHeight int64 = 47
	for i := startHeight; i <= totalHeight; i++ {
		valInfo := oldver.LoadOldValidatorsInfo(stateDbm, i)
		if valInfo != nil {
			res, _ := json.Marshal(valInfo)
			t.Log(string(res))
		}
	}
}

func TestLoadOldGenesisDoc(t *testing.T) {
	onTestStart()
	defer onTestStop()

	genDoc, err := oldver.LoadOldGenesisDoc(stateDbm)
	if err == nil {
		res, _ := json.Marshal(genDoc)
		t.Log(string(res))
	}
}

func TestLoadOldState(t *testing.T) {
	onTestStart()
	defer onTestStop()

	state := oldver.LoadOldState(stateDbm)
	if &state != nil {
		res, _ := json.Marshal(state)
		t.Log(string(res))
	}
}

package convert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/commis/tm-upgrade/util"

	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func OnConfigToml(configFilePath string) {
	var configTmpl = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

proxy_app = "tcp://0.0.0.0:46658"
moniker = "anonymous"
node_laddr = "tcp://0.0.0.0:46656"
seeds = ""
fast_sync = true
db_backend = "leveldb"
log_level = "info"
rpc_laddr = "tcp://0.0.0.0:46657"
`
	cmn.WriteFile(configFilePath, []byte(configTmpl), 0644)
}

func UpgradeGenesisJSON(ldb db.DB, oPath, nPath string) {
	jsonBytes, err := ioutil.ReadFile(oPath)
	if err != nil {
		cmn.Exit(err.Error())
	}

	oGen, err := his.GenesisDocFromJSON(jsonBytes)
	if err == nil {
		nGen := NewGenesisDoc(oGen)
		if err := nGen.SaveAs(nPath); err != nil {
			cmn.Exit(err.Error())
		}

		if err := nGen.ValidateAndComplete(); err == nil {
			SaveNewGenesisDoc(ldb, nGen)
		}
	}
}

func NewPrivValidator(oPath string, privVali *privval.FilePV) {
	old := his.LoadPrivValidator(oPath)
	privVali.Address = old.Address.Bytes()
	privVali.LastHeight = old.LastHeight
	privVali.LastRound = old.LastRound
	privVali.LastStep = old.LastStep
	privVali.LastSignature = CvtNewSignature(old.LastSignature)
	privVali.LastSignBytes = old.LastSignBytes.Bytes()
	privVali.PubKey = CvtNewPubKey(old.PubKey)
	privVali.PrivKey = CvtNewPrivKey(old.PrivKey)
}

func SaveNewGenesisDoc(ldb db.DB, genDoc *types.GenesisDoc) {
	bytes, err := json.Marshal(genDoc)
	if err != nil {
		fmt.Printf("Failed to save genesis doc due to marshaling error: %v", err)
		return
	}
	ldb.Set([]byte(util.GenesisDocKey), bytes)
}

func LoadOldGenesisDoc(db dbm.DB) *his.GenesisDoc {
	bytes := db.Get([]byte(util.GenesisDocKey))
	if len(bytes) == 0 {
		return nil
	}

	var genDoc *his.GenesisDoc
	err := json.Unmarshal(bytes, &genDoc)
	if err != nil {
		log.Printf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, bytes)
	}
	return genDoc
}

func NewGenesisDoc(old *his.GenesisDoc) *types.GenesisDoc {
	newGenesisDoc := &types.GenesisDoc{
		AppHash:     old.AppHash.Bytes(),
		ChainID:     old.ChainID,
		GenesisTime: old.GenesisTime,
		Validators:  []types.GenesisValidator{},
	}
	for _, val := range old.Validators {
		one := types.GenesisValidator{}
		one.Power = val.Power
		one.Name = val.Name
		one.PubKey = CvtNewPubKey(val.PubKey)

		newGenesisDoc.Validators = append(newGenesisDoc.Validators, one)
	}

	return newGenesisDoc
}

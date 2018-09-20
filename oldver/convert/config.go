package convert

import (
	"io/ioutil"

	"github.com/commis/tm-tools/oldver"
	his "github.com/commis/tm-tools/oldver/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/privval"
	cmn "github.com/tendermint/tmlibs/common"
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
		panic(err)
	}

	oGen, err := his.GenesisDocFromJSON(jsonBytes)
	if err == nil {
		nGen := NewGenesisDoc(oGen)
		if err := nGen.SaveAs(nPath); err != nil {
			panic(err)
		}

		if err := nGen.ValidateAndComplete(); err == nil {
			oldver.SaveNewGenesisDoc(ldb, nGen)
		}
	}
}

func NewPrivValidator(oPath string, privVali *privval.FilePV) {
	old := his.LoadPrivValidator(oPath)
	privVali.Address = old.Address.Bytes()
	privVali.LastHeight = old.LastHeight
	privVali.LastRound = old.LastRound
	privVali.LastStep = old.LastStep
	privVali.LastSignature = NewSignature(old.LastSignature)
	privVali.LastSignBytes = old.LastSignBytes.Bytes()
	privVali.PubKey = NewPubKey(old.PubKey)
	privVali.PrivKey = NewPrivKey(old.PrivKey)
}

package commands

import (
	"github.com/commis/tm-tools/libs/op"
	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tmlibs/common"
)

type MigrateParam struct {
	oldData string
	newData string
	height  int64
}

var mp = &MigrateParam{}

func init() {
	MigrateDataCmd.Flags().StringVar(&mp.oldData, "old", "/home/tendermint", "Directory of old tendermint data")
	MigrateDataCmd.Flags().StringVar(&mp.newData, "new", "/home/tendermint.new", "Directory of new tendermint data")
	MigrateDataCmd.Flags().Int64Var(&mp.height, "h", 1, "Migrate the start block height")
}

var MigrateDataCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Tendermint data migrator",
	RunE:  migrateData,
}

func migrateData(cmd *cobra.Command, args []string) error {
	cmn.EnsureDir(mp.newData+"/config", 0755)
	cmn.EnsureDir(mp.newData+"/data", 0755)

	convert := op.TmDataStore{}
	convert.OnStart(mp.oldData, mp.newData)
	defer convert.OnStop()

	// convert config
	convert.OnGenesisJSON(mp.oldData+"/config/genesis.json", mp.newData+"/config/genesis.json")
	convert.OnPrivValidatorJSON(mp.oldData+"/config/priv_validator.json", mp.newData+"/config/priv_validator.json")

	// convert data
	convert.TotalHeight(false)
	convert.OnBlockStore(mp.height)
	convert.OnCsWal(mp.oldData+"/data", mp.newData+"/data")
	convert.OnEvidence()

	return nil
}

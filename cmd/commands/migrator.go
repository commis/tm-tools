package commands

import (
	cvt "github.com/commis/tm-tools/libs/convert"
	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tmlibs/common"
)

var (
	oldData string
	newData string
	sHeight int64
)

func init() {
	MigrateDataCmd.Flags().StringVar(&oldData, "old", "/home/tendermint", "Directory of old tendermint data")
	MigrateDataCmd.Flags().StringVar(&newData, "new", "/home/tendermint.new", "Directory of new tendermint data")
	MigrateDataCmd.Flags().Int64Var(&sHeight, "h", 1, "Migrate the start block height")
}

var MigrateDataCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Tendermint data migrator",
	RunE:  migrateData,
}

func migrateData(cmd *cobra.Command, args []string) error {
	cmn.EnsureDir(newData+"/config", 0755)
	cmn.EnsureDir(newData+"/data", 0755)

	convert := cvt.Converter{}
	convert.OnStart(oldData, newData)
	defer convert.OnStop()

	// convert config
	//convert.OnConfigToml(nTmRoot + "/config.toml")
	convert.OnGenesisJSON(oldData+"/config/genesis.json", newData+"/config/genesis.json")
	convert.OnPrivValidatorJSON(oldData+"/config/priv_validator.json", newData+"/config/priv_validator.json")

	// convert data
	convert.TotalHeight()
	convert.OnBlockStore(sHeight)
	convert.OnEvidence()

	return nil
}

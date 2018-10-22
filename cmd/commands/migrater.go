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

	//01.upgrade all genesis.json and priv_validator.json
	tmCfg := op.CreateTmConfig(mp.oldData, mp.newData)
	tmCfg.UpgradeAllGenesisAndPvFile()
	defer tmCfg.UpdateAllPrivValFile()

	//02.upgrade all db data
	needUpgrade := false
	func() {
		tmStore := op.CreateTmDataStore(mp.oldData, mp.newData)
		defer tmStore.OnStop()

		// tmStore data
		tmStore.GetTotalHeight(false)
		if needUpgrade = tmStore.CheckNeedUpgrade(tmCfg.GetTopPrivVal()); needUpgrade {
			tmStore.OnBlockStore(mp.height)
		} /*else {
			tmStore.UpdateGenesisDocInStateDB()
		}*/
	}()

	//03.upgrade cs.wal file
	if needUpgrade {
		op.UpgradeNodeCsWal(mp.oldData, mp.newData)
	}

	return nil
}

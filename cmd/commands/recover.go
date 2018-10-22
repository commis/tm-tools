package commands

import (
	"fmt"

	"github.com/commis/tm-tools/libs/op"
	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type RecoverParam struct {
	dataPath string
	ver      bool
	height   int64
}

var rp = &RecoverParam{}

func init() {
	ResetBlockCmd.Flags().StringVar(&rp.dataPath, "db", "/home/tendermint", "Directory of tendermint data")
	ResetBlockCmd.Flags().BoolVar(&rp.ver, "v", false, "Whether new version data")
	ResetBlockCmd.Flags().Int64Var(&rp.height, "h", 0, "Recover block height")
}

var ResetBlockCmd = &cobra.Command{
	Use:   "recover",
	Short: "Tendermint block recover",
	RunE:  resetBlockHeight,
}

func resetBlockHeight(cmd *cobra.Command, args []string) error {
	if rp.height == 0 {
		cmn.Exit(fmt.Sprint("recover height is not setting"))
	}

	//01.recover ethereum height
	op.ResetEthHeight(rp.dataPath, rp.height)

	tdStore := op.CreateTmDataStore(mp.oldData, mp.newData)
	defer tdStore.OnStop()

	var tmDataStore *op.TmDataStore
	if rp.ver {
		op.ResetPrivValHeight(op.TMVer0180, rp.dataPath, rp.height)
		op.ResetNodeWalHeight(op.TMVer0180, rp.dataPath, rp.height)

		tmDataStore = op.CreateTmDataStore("", rp.dataPath)
	} else {
		op.ResetPrivValHeight(op.TMVer0231, rp.dataPath, rp.height)
		op.ResetNodeWalHeight(op.TMVer0231, rp.dataPath, rp.height)

		tmDataStore = op.CreateTmDataStore(rp.dataPath, "")
	}
	defer tmDataStore.OnStop()

	// recover data
	tmDataStore.GetTotalHeight(rp.ver)
	tmDataStore.OnBlockRecover(rp.ver, rp.height)

	return nil
}

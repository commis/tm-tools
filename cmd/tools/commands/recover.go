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
	RunE:  resetBlockHight,
}

func resetBlockHight(cmd *cobra.Command, args []string) error {
	if rp.height == 0 {
		cmn.Exit(fmt.Sprint("recover height is not setting"))
	}

	reset := op.TmDataStore{}
	if rp.ver {
		reset.OnStart("", rp.dataPath)
	} else {
		reset.OnStart(rp.dataPath, "")
	}
	defer reset.OnStop()

	// recover data
	reset.TotalHeight(rp.ver)
	reset.OnBlockRecover(rp.ver, rp.height)
	reset.OnEvidenceRecover(rp.ver, rp.height)

	return nil
}

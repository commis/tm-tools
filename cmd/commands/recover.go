package commands

import (
	"fmt"

	"github.com/commis/tm-tools/libs/op"
	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	dataPath string
	dbVer    bool
	eHeight  int64
)

func init() {
	ResetBlockCmd.Flags().StringVar(&dataPath, "db", "/home/tendermint", "Directory of tendermint data")
	ResetBlockCmd.Flags().BoolVar(&dbVer, "v", false, "Whether new version data")
	ResetBlockCmd.Flags().Int64Var(&eHeight, "h", 0, "Recover block height")
}

var ResetBlockCmd = &cobra.Command{
	Use:   "recover",
	Short: "Tendermint block recover",
	RunE:  resetBlockHight,
}

func resetBlockHight(cmd *cobra.Command, args []string) error {
	if eHeight == 0 {
		cmn.Exit(fmt.Sprint("recover height is not setting"))
	}

	reset := op.TmDataStore{}
	if dbVer {
		reset.OnStart("", dataPath)
	} else {
		reset.OnStart(dataPath, "")
	}
	defer reset.OnStop()

	// recover data
	reset.TotalHeight(dbVer)
	reset.OnBlockRecover(dbVer, eHeight)
	reset.OnEvidenceRecover(dbVer, eHeight)

	return nil
}

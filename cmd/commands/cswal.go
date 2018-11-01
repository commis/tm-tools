package commands

import (
	"github.com/commis/tm-tools/libs/op"
	"github.com/spf13/cobra"
)

type CsWalParameter struct {
	tmPath string
	ver    bool
	init   bool
}

var cw = &CsWalParameter{}

func init() {
	ViewWalCmd.Flags().StringVar(&cw.tmPath, "p", "/home/tendermint", "tendermint path")
	ViewWalCmd.Flags().BoolVar(&cw.ver, "v", false, "Whether new version data")
	ViewWalCmd.Flags().BoolVar(&cw.init, "init", false, "update vote in priv_validator.json")
}

var ViewWalCmd = &cobra.Command{
	Use:   "cswal",
	Short: "Tendermint cs.wal viewer",
	RunE:  showCsWal,
}

func showCsWal(cmd *cobra.Command, args []string) error {
	if !cw.ver {
		if cw.init {
			op.InitWalMessage(op.TMVer0180, cw.tmPath)
			return nil
		}
		op.PrintWalMessage(op.TMVer0180, cw.tmPath)
	} else {
		if cw.init {
			op.InitWalMessage(op.TMVer0231, cw.tmPath)
			return nil
		}
		op.PrintWalMessage(op.TMVer0231, cw.tmPath)
	}

	return nil
}

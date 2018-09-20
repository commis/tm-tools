package commands

import "github.com/spf13/cobra"

var ResetBlockCmd = &cobra.Command{
	Use:   "reset",
	Short: "Tendermint block reset",
	RunE:  resetBlockHight,
}

func resetBlockHight(cmd *cobra.Command, args []string) error {
	//return initFilesWithConfig(config)
	return nil
}

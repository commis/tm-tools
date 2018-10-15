package main

import (
	cmd "github.com/commis/tm-tools/cmd/tools/commands"
	"github.com/commis/tm-tools/libs/cli"
)

func main() {
	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		cmd.ViewDatabaseCmd,
		cmd.MigrateDataCmd,
		cmd.ResetBlockCmd,
		cmd.VersionCmd)

	cmd := cli.PrepareBaseCmd(rootCmd)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

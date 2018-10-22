package commands

import (
	"github.com/commis/tm-tools/libs/cli"
	"github.com/commis/tm-tools/libs/log"
	"github.com/spf13/cobra"
)

var (
	config = cli.DefaultConfig()
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().Int("log_level", config.LogLevel, "Log level (Trace/Debug/Info/Warn/Error)=(0/1/2/3/4)")
}

var RootCmd = &cobra.Command{
	Use:   "tm_tools",
	Short: "Tendermint upgrade tools in Go",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		log.Log.SetDebugLevel(cli.DefaultLogLevel())

		return nil
	},
}

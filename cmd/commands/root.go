package commands

import (
	cfg "github.com/commis/tm-tools/config"
	"github.com/commis/tm-tools/libs/log"
	"github.com/spf13/cobra"
)

var (
	config = cfg.DefaultConfig()
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

		log.Log.SetDebugLevel(cfg.DefaultLogLevel())

		return nil
	},
}

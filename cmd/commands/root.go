package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/commis/tm-tools/config"
	"github.com/commis/tm-tools/libs/cli"
	"github.com/commis/tm-tools/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	config = cfg.DefaultConfig()
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log_level", config.LogLevel, "Log level")
}

var RootCmd = &cobra.Command{
	Use:   "tm_tools",
	Short: "Tendermint upgrade tools in Go",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		logger, err = flags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		return nil
	},
}

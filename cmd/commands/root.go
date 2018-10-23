package commands

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/commis/tm-tools/libs/cli"
	"github.com/commis/tm-tools/libs/log"
	"github.com/spf13/cobra"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	debugPort = 39999
	config    = cli.DefaultConfig()
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

		startPerformanceTracePort()

		return nil
	},
}

func startPerformanceTracePort() {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", debugPort), nil)
		if err != nil {
			cmn.Exit(fmt.Sprintf("failed to listen debug port: %v", err))
		}
		log.Infof("start performance trace port: %d", debugPort)
	}()
}

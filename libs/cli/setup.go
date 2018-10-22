package cli

import (
	"os"

	"github.com/commis/tm-tools/libs/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	TraceFlag = "trace"
)

type Executable interface {
	Execute() error
}

type Executor struct {
	*cobra.Command
	Exit func(int) // this is os.Exit by default, override in tests
}

func PrepareBaseCmd(cmd *cobra.Command) Executor {
	//cobra.OnInitialize(func() { initEnv(envPrefix) })
	cmd.PersistentFlags().Bool(TraceFlag, true, "print out full stack trace on errors")
	return Executor{cmd, os.Exit}
}

type ExitCoder interface {
	ExitCode() int
}

func (e Executor) Execute() error {
	e.SilenceUsage = true
	e.SilenceErrors = true
	err := e.Command.Execute()
	if err != nil {
		if viper.GetBool(TraceFlag) {
			log.Infof("ERROR: %+v", err)
		} else {
			log.Infof("ERROR: %v", err)
		}

		// return error code 1 by default, can override it with a special error type
		exitCode := 1
		if ec, ok := err.(ExitCoder); ok {
			exitCode = ec.ExitCode()
		}
		e.Exit(exitCode)
	}
	return err
}

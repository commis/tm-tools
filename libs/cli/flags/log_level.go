package flags

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
)

// Example:
//		ParseLogLevel("debug", log.NewTMLogger(os.Stdout), "error|debug|info")
func ParseLogLevel(lvl string, logger log.Logger, defaultLogLevelValue string) (log.Logger, error) {
	var level string
	if lvl == "" {
		level = defaultLogLevelValue
	} else {
		level = lvl
	}

	options := make([]log.Option, 0)
	option, err := log.AllowLevel(level)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to parse default log level (%s)", lvl))
	}
	options = append(options, option)

	return log.NewFilter(logger, options...), nil
}

package flags_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/commis/tm-tools/libs/cli/flags"

	"github.com/tendermint/tendermint/libs/log"
)

const (
	defaultLogLevelValue = "info"
)

func TestParseLogLevel(t *testing.T) {
	var buf bytes.Buffer
	jsonLogger := log.NewTMJSONLogger(&buf)

	correctLogLevels := []struct {
		lvl              string
		expectedLogLines []string
	}{
		{"error", []string{
			``,
			``,
			`{"_msg":"Mesmero","level":"error","module":"view"}`,
			``}},
		{"debug", []string{
			`{"_msg":"Gideon","level":"debug","module":"wire"}`,
			`{"_msg":"Mind","level":"info","module":"state"}`,
			`{"_msg":"Mesmero","level":"error","module":"view"}`,
			`{"_msg":"Kitty Pryde","level":"debug"}`}},
		{"info", []string{
			``,
			`{"_msg":"Mind","level":"info","module":"state"}`,
			`{"_msg":"Mesmero","level":"error","module":"view"}`,
			``}},
		{"", []string{
			``,
			`{"_msg":"Mind","level":"info","module":"state"}`,
			`{"_msg":"Mesmero","level":"error","module":"view"}`,
			``}},
	}

	for _, c := range correctLogLevels {
		logger, err := flags.ParseLogLevel(c.lvl, jsonLogger, defaultLogLevelValue)
		if err != nil {
			t.Fatal(err)
		}

		buf.Reset()
		logger.With("module", "wire").Debug("Gideon")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[0] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[0], have, c.lvl)
		}

		buf.Reset()
		logger.With("module", "state").Info("Mind")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[1] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[1], have, c.lvl)
		}

		buf.Reset()
		logger.With("module", "view").Error("Mesmero")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[2] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[2], have, c.lvl)
		}

		buf.Reset()
		logger.Debug("Kitty Pryde")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[3] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[3], have, c.lvl)
		}
	}
}

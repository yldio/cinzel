// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A configuration warning quotes the key it read out of .cinzelrc.yaml, and a
// key is free to carry an ANSI escape sequence. Written to a terminal as it
// stands, the sequence is acted on rather than shown: a crafted key erases the
// warning naming it and leaves a line of its own in its place. The errors the
// tool ends on are escaped for the same reason.
func TestAConfigWarningDoesNotCarryAnEscapeSequence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
	}{
		{
			name:   "an unknown command key",
			config: "github:\n  \"x\\e[2K\\rparsed with no warnings\": 1\n",
		},
		{
			name:   "an unknown key under a command",
			config: "github:\n  parse:\n    \"x\\e[2K\\rparsed with no warnings\": 1\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeFile(t, configFilename, []byte(tc.config))

				app, errOut, p := newConfigTestApp(t)

				if err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p}); err != nil {
					t.Fatalf("Execute() error = %v", err)
				}

				warnings := errOut.String()

				if warnings == "" {
					t.Fatal("want a warning, got none")
				}

				if strings.ContainsAny(warnings, "\x1b\r") {
					t.Errorf("a control character reached the output: %q", warnings)
				}
			})
		})
	}
}

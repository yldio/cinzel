// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"strings"
	"testing"
)

// pin and upgrade choose PinFile or PinDirectory from which flag was set, not
// from what the path turns out to be, so --file pointed at a directory walked
// it. Parse and unparse refuse that now, and a flag cannot mean a file in one
// command and a directory in the next.
func TestPinAndUpgradeRefuseADirectoryAsFile(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{
			name: "pin",
			args: []string{"cinzel", "github", "pin", "--file", "actions", "--dry-run"},
		},
		{
			name: "upgrade",
			args: []string{"cinzel", "github", "upgrade", "--file", "actions", "--dry-run"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeActionDir(t, pinnedAction)

				err := runCommandLine(t, tc.args)
				if err == nil {
					t.Fatal("--file pointed at a directory was accepted")
				}

				if !strings.Contains(err.Error(), "names a directory") {
					t.Errorf("error = %v, want it to name the directory problem", err)
				}
			})
		})
	}
}

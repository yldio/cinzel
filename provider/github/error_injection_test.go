// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/provider"
)

// Validation errors quote the job name they came from, and a YAML file can
// spell one with an escape sequence. The message used to carry the escape to
// the terminal, where it was acted on rather than shown, so every message the
// unparse path produces for a crafted name has to come back inert.
func TestValidationErrorsCarryNoTerminalEscape(t *testing.T) {
	// A raw escape byte is refused by the YAML decoder as a control
	// character, so the name carries the escape as a YAML escape sequence,
	// which the decoder turns back into the byte.
	const escaped = `"\u001b[31mPWNED\u001b[0m"`

	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "job missing runs-on",
			yaml: "on: push\njobs:\n  " + escaped + ":\n    steps:\n      - run: echo hi\n",
		},
		{
			name: "job with invalid permissions",
			yaml: "on: push\njobs:\n  " + escaped + ":\n    runs-on: ubuntu-latest\n    permissions: bogus\n    steps:\n      - run: echo hi\n",
		},
		{
			name: "step with an unversioned uses",
			yaml: "on: push\njobs:\n  " + escaped + ":\n    runs-on: ubuntu-latest\n    steps:\n      - uses: \"no version here\"\n",
		},
		{
			name: "needs naming a job that does not exist",
			yaml: "on: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    needs: [" + escaped + "]\n    steps:\n      - run: echo hi\n",
		},
		{
			name: "job that is not a mapping",
			yaml: "on: push\njobs:\n  " + escaped + ": not-an-object\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "workflow.yaml")

			if err := os.WriteFile(path, []byte(tc.yaml), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Unparse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})
			if err == nil {
				t.Fatal("expected the crafted workflow to be rejected")
			}

			// The CLI prints the error through this same call, so the check
			// covers what a user actually sees.
			message := cinzelerror.SafeForTerminal(cinzelerror.New(err).Err.Error())

			if strings.ContainsRune(message, rune(0x1b)) {
				t.Errorf("escape reached the message: %q", message)
			}

			if !strings.Contains(message, "PWNED") {
				t.Errorf("message does not quote the name at all, so the case proves nothing: %q", message)
			}
		})
	}
}

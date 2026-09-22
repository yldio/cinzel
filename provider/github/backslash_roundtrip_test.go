// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A "run" script becomes an HCL heredoc, which has no escapes at all and so
// writes a backslash single. That is the parity unescape.Unicode read as an
// escape of its own, so the six characters the author typed were replaced by
// the one they name, in the middle of a shell command, at exit 0.
func TestABackslashInARunSurvivesTheRoundtrip(t *testing.T) {
	const yml = `name: CI
on:
  push:
    branches: [main]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Test
        run: |
          printf X\u00e9Y
          echo done
`

	hcl, back := roundtripWorkflowYAML(t, yml)

	const want = `printf X\u00e9Y`

	if !strings.Contains(hcl, want) {
		t.Errorf("want %q in HCL, got:\n%s", want, hcl)
	}

	if !strings.Contains(back, want) {
		t.Errorf("want %q in YAML, got:\n%s", want, back)
	}
}

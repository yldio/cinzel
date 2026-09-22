// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A warning quotes a key read out of the YAML, and a key is free to carry an
// ANSI escape sequence. Written to a terminal as it stands, the sequence is
// acted on rather than shown: a crafted key erases the warning naming it and
// leaves a line of its own in its place, so a run that dropped a key reads as
// a run that dropped nothing. The errors the tool ends on are escaped for the
// same reason.
func TestADroppedKeyWarningDoesNotCarryAnEscapeSequence(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
	}{
		{
			name: "a workflow key",
			yml:  "workflow:\n  name: pipe\n  \"x\\e[2K\\rPipeline written with no warnings\": 1\n",
		},
		{
			name: "a top-level key",
			yml:  "\"x\\e[2K\\rPipeline written with no warnings\": 1\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yml := tc.yml + "build:\n  script:\n    - echo hi\n"

			// A key of this shape is also refused as an HCL identifier, and
			// that error is escaped where it is printed. The warning is
			// written before it, so the run failing does not settle this.
			warnings := captureStderr(t, func() {
				_ = unparseYAMLString(t, yml)
			})

			if warnings == "" {
				t.Fatal("want a warning, got none")
			}

			if strings.ContainsAny(warnings, "\x1b\r") {
				t.Errorf("a control character reached the output: %q", warnings)
			}
		})
	}
}

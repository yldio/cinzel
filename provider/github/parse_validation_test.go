// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// validateParsedWorkflow used to run before "jobs" was assigned, so every
// check that walks jobs and steps saw an empty workflow and was dead on the
// parse path. These go through Parse rather than calling the validator, which
// is the part that was not happening.
func TestParseValidatesInsideJobs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		step    string
		wantErr string
	}{
		{
			name:    "unclosed expression in a run",
			step:    `  run = "echo $${{ github.ref "`,
			wantErr: "unclosed expression",
		},
		{
			name:    "unclosed expression in an env value",
			step:    "  run = \"echo hi\"\n\n  env {\n    name  = \"URL\"\n    value = \"$${{ broken\"\n  }",
			wantErr: "unclosed expression",
		},
		{
			name: "a sound step is unaffected",
			step: `  run = "echo $${{ github.ref }}"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := "workflow \"in\" {\n  filename = \"in\"\n  on \"push\" {}\n  jobs = [job.a]\n}\n\n" +
				"job \"a\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.s]\n}\n\n" +
				"step \"s\" {\n" + tc.step + "\n}\n"

			dir := t.TempDir()
			path := filepath.Join(dir, "in.hcl")

			if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("want an error, got nil")
			}

			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want %q in %q", tc.wantErr, err)
			}

			// The path has to name where in the document it is, or a large
			// workflow gives nothing to go on.
			if !strings.Contains(err.Error(), "jobs.a.steps[0]") {
				t.Errorf("want the step path in %q", err)
			}
		})
	}
}

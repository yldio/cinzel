// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// Two jobs reaching the same YAML key used to overwrite one another in the
// map. The file held whichever came last and the command exited 0.
func TestTwoJobsCannotShareAKey(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hcl   string
		want  error
		named []string
	}{
		{
			name:  "the same id attribute",
			hcl:   twoJobsHCL("first", "second", `  id = "same"`),
			want:  errDuplicateJobKey,
			named: []string{"first", "second", "same"},
		},
		{
			name: "distinct jobs are unaffected",
			hcl:  twoJobsHCL("first", "second", ""),
			want: nil,
		},
		{
			name:  "the same block label",
			hcl:   twoJobsHCL("build", "build", ""),
			want:  errDuplicateJobLabel,
			named: []string{"build"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "in.hcl")

			if err := os.WriteFile(path, []byte(tc.hcl), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})

			if tc.want == nil {
				if err != nil {
					t.Fatalf("distinct jobs must parse, got %v", err)
				}

				return
			}

			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}

			// The message has to name what collided, or there is nothing to
			// go and fix.
			for _, want := range tc.named {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("want %q named in %q", want, err)
				}
			}
		})
	}
}

// The same key in two workflows is not a collision: each is its own file.
func TestTheSameJobKeyInTwoWorkflowsIsFine(t *testing.T) {
	hcl := "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		jobBlock("first", `  id = "same"`) +
		jobBlock("second", `  id = "same"`) +
		"workflow \"one\" {\n  filename = \"one\"\n  on \"push\" {}\n  jobs = [job.first]\n}\n\n" +
		"workflow \"two\" {\n  filename = \"two\"\n  on \"push\" {}\n  jobs = [job.second]\n}\n"

	out := parseWorkflow(t, hcl)

	for _, want := range []string{"one.yaml", "two.yaml"} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("want %s written, got %v", want, err)
		}
	}
}

// twoJobsHCL returns a workflow running two jobs with the given labels, each
// carrying the same extra attribute line.
func twoJobsHCL(first, second, extra string) string {
	refs := "job." + first

	if first != second {
		refs += ", job." + second
	}

	return "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		jobBlock(first, extra) +
		jobBlock(second, extra) +
		"workflow \"w\" {\n" +
		"  filename = \"w\"\n" +
		"  on \"push\" {}\n" +
		"  jobs = [" + refs + "]\n" +
		"}\n"
}

func jobBlock(id, extra string) string {
	block := "job \"" + id + "\" {\n"

	if extra != "" {
		block += extra + "\n"
	}

	return block +
		"  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		"  steps = [step.hi]\n" +
		"}\n\n"
}

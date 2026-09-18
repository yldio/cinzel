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

// GitHub requires a step id to be unique within its job. Two step blocks
// reaching the same id through their own "id" attributes went out as a written
// file and exit 0, and only actionlint or GitHub itself said otherwise.
func TestTwoStepsInOneJobCannotShareAnID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hcl   string
		want  error
		named []string
	}{
		{
			name:  "the same id attribute",
			hcl:   oneJobTwoStepsHCL(`  id = "same"`),
			want:  errDuplicateStepID,
			named: []string{"first", "second", "same"},
		},
		{
			name: "distinct steps are unaffected",
			hcl:  oneJobTwoStepsHCL(""),
			want: nil,
		},
		{
			// Neither writes an id at all, so there is nothing to collide.
			name: "ignore_id on both",
			hcl:  oneJobTwoStepsHCL("  ignore_id = true"),
			want: nil,
		},
		{
			// A step written under a comment arrives wrapped, and the id was
			// read by asserting the map straight off the value. Every
			// commented step read as having no id, so two of them could write
			// the same one.
			name:  "the same id under a comment",
			hcl:   commentedStepsHCL(`  id = "same"`),
			want:  errDuplicateStepID,
			named: []string{"first", "second", "same"},
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
					t.Fatalf("distinct steps must parse, got %v", err)
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

// The same step id in two jobs is not a collision: GitHub scopes a step id to
// its own job.
func TestTheSameStepIDInTwoJobsIsFine(t *testing.T) {
	hcl := stepBlock("first", `  id = "same"`) +
		stepBlock("second", `  id = "same"`) +
		"job \"build\" {\n  runs_on { runners = \"ubuntu-latest\" }\n  steps = [step.first]\n}\n\n" +
		"job \"test\" {\n  runs_on { runners = \"ubuntu-latest\" }\n  steps = [step.second]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.build, job.test]\n}\n"

	out := parseWorkflow(t, hcl)

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(string(content), "id: same"); got != 2 {
		t.Errorf("want both jobs to keep the id, got %d in:\n%s", got, content)
	}
}

// A job running the same step twice keeps the id on the first occurrence only,
// which is what makes the run legal. The check must not mistake that for a
// collision.
func TestAStepRepeatedInOneJobStillParses(t *testing.T) {
	hcl := stepBlock("only", "") +
		"job \"build\" {\n  runs_on { runners = \"ubuntu-latest\" }\n  steps = [step.only, step.only]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.build]\n}\n"

	out := parseWorkflow(t, hcl)

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(string(content), "id: only"); got != 1 {
		t.Errorf("want the id on the first occurrence only, got %d in:\n%s", got, content)
	}
}

// stepBlock returns a step block with the given label and extra attribute line.
func stepBlock(label, extra string) string {
	block := "step \"" + label + "\" {\n  run = \"echo hi\"\n"

	if extra != "" {
		block += extra + "\n"
	}

	return block + "}\n\n"
}

// commentedStepsHCL is oneJobTwoStepsHCL with a comment above each step, which
// wraps the converted step and used to hide its id.
func commentedStepsHCL(extra string) string {
	return "// the first one\n" + stepBlock("first", extra) +
		"// the second one\n" + stepBlock("second", extra) +
		"job \"build\" {\n  runs_on { runners = \"ubuntu-latest\" }\n  steps = [step.first, step.second]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.build]\n}\n"
}

// oneJobTwoStepsHCL returns a workflow whose single job runs two steps, each
// carrying the same extra attribute line.
func oneJobTwoStepsHCL(extra string) string {
	return stepBlock("first", extra) +
		stepBlock("second", extra) +
		"job \"build\" {\n  runs_on { runners = \"ubuntu-latest\" }\n  steps = [step.first, step.second]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.build]\n}\n"
}

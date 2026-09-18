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

// GitHub scopes a step id to the composite action running it, exactly as it
// does to a job. The action path had no uniqueness guard at all, so two steps
// writing the same id went out as a written action.yml and exit 0.
func TestTwoStepsInOneActionCannotShareAnID(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hcl   string
		want  error
		named []string
	}{
		{
			name:  "the same id attribute",
			hcl:   oneActionTwoStepsHCL(`  id = "same"`),
			want:  errDuplicateActionStepID,
			named: []string{"first", "second", "same"},
		},
		{
			name: "distinct steps are unaffected",
			hcl:  oneActionTwoStepsHCL(""),
			want: nil,
		},
		{
			// Neither writes an id at all, so there is nothing to collide.
			name: "ignore_id on both",
			hcl:  oneActionTwoStepsHCL("  ignore_id = true"),
			want: nil,
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

// An action running the same step twice keeps the id on the first occurrence
// only, which is what makes the run legal. The guard must not read that as a
// collision.
func TestAStepRepeatedInOneActionStillParses(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	path := filepath.Join(dir, "in.hcl")

	hcl := stepBlock("only", "") +
		compositeActionHCL("alpha", "steps = [step.only, step.only]")

	if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(out, "alpha", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(string(content), "id: only"); got != 1 {
		t.Errorf("want the id on the first occurrence only, got %d in:\n%s", got, content)
	}
}

// oneActionTwoStepsHCL returns a composite action whose two steps each carry
// the same extra attribute line.
func oneActionTwoStepsHCL(extra string) string {
	return stepBlock("first", extra) +
		stepBlock("second", extra) +
		compositeActionHCL("alpha", "steps = [step.first, step.second]")
}

// compositeActionHCL returns a composite action with the given runs body line.
func compositeActionHCL(id, runs string) string {
	return "action \"" + id + "\" {\n" +
		"  filename = \"" + id + "\"\n" +
		"  name = \"A\"\n" +
		"  description = \"d\"\n" +
		"  runs {\n    using = \"composite\"\n    " + runs + "\n  }\n}\n"
}

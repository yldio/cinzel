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

// A step label is the address step.<label> resolves. Two blocks claiming one
// left the reference ambiguous, and the message named neither block.
func TestTwoStepBlocksCannotShareALabel(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files map[string]string
		want  error
		named []string
	}{
		{
			name:  "in one file",
			files: map[string]string{"in.hcl": labelStepBlock("build", "one") + labelStepBlock("build", "two") + labelJobAndWorkflow("build")},
			want:  errDuplicateStepLabel,
			named: []string{"build", "in.hcl:1,", "in.hcl:5,"},
		},
		{
			name: "across two files in one directory",
			files: map[string]string{
				"a.hcl": labelStepBlock("build", "one") + labelJobAndWorkflow("build"),
				"b.hcl": labelStepBlock("build", "two"),
			},
			want:  errDuplicateStepLabel,
			named: []string{"build", "a.hcl:1,", "b.hcl:1,"},
		},
		{
			name: "a duplicate nothing references still collides",
			files: map[string]string{
				"in.hcl": labelStepBlock("used", "make") + labelStepBlock("orphan", "one") +
					labelStepBlock("orphan", "two") + labelJobAndWorkflow("used"),
			},
			want:  errDuplicateStepLabel,
			named: []string{"orphan"},
		},
		{
			name:  "distinct labels are unaffected",
			files: map[string]string{"in.hcl": labelStepBlock("first", "one") + labelStepBlock("second", "two") + labelJobAndWorkflow("first")},
			want:  nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			for name, body := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			err := New().Parse(provider.ProviderOps{Directory: dir, OutputDirectory: filepath.Join(dir, "out")})

			if tc.want == nil {
				if err != nil {
					t.Fatalf("distinct labels must parse, got %v", err)
				}

				return
			}

			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}

			// Without a file and a line the author has the whole directory to
			// search, which is what this error used to leave them with.
			for _, want := range tc.named {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("want %q named in %q", want, err)
				}
			}
		})
	}
}

// Two directories are two parses, so a label repeated between them is not a
// collision unless the run is recursive.
func TestTheSameStepLabelInTwoDirectoriesIsFine(t *testing.T) {
	root := t.TempDir()

	for _, sub := range []string{"one", "two"} {
		dir := filepath.Join(root, sub)

		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}

		body := labelStepBlock("build", sub) +
			"job \"j_" + sub + "\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.build]\n}\n\n" +
			"workflow \"w_" + sub + "\" {\n  filename = \"" + sub + "\"\n  on \"push\" {}\n  jobs = [job.j_" + sub + "]\n}\n"

		if err := os.WriteFile(filepath.Join(dir, "in.hcl"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}

		out := filepath.Join(root, "out", sub)

		if err := New().Parse(provider.ProviderOps{Directory: dir, OutputDirectory: out}); err != nil {
			t.Fatalf("%s must parse on its own, got %v", sub, err)
		}

		if _, err := os.Stat(filepath.Join(out, sub+".yaml")); err != nil {
			t.Errorf("want %s.yaml written, got %v", sub, err)
		}
	}
}

func labelStepBlock(label, run string) string {
	return "step \"" + label + "\" {\n  run = \"" + run + "\"\n}\n\n"
}

func labelJobAndWorkflow(label string) string {
	return "job \"j\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step." + label + "]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.j]\n}\n"
}

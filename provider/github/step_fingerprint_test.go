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

// One block stands for every step with the same content. The fingerprint used
// to drop "id" first, so two steps differing only by id merged and the second
// id was gone, leaving a "steps.<id>.outputs" reference naming nothing.
func TestStepsDifferingOnlyByIDStaySeparate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		steps     string
		wantSteps []string
		wantGone  []string
	}{
		{
			name:      "two explicit ids on the same content",
			steps:     "      - id: first\n        run: make thing\n      - id: second\n        run: make thing\n",
			wantSteps: []string{`step "first"`, `step "second"`},
		},
		{
			name:      "no id at all still dedupes",
			steps:     "      - run: make thing\n      - run: make thing\n",
			wantSteps: []string{`step "make"`},
			wantGone:  []string{`step "make_2"`},
		},
		{
			name:      "one id and one without are not the same step",
			steps:     "      - id: named\n        run: make thing\n      - run: make thing\n",
			wantSteps: []string{`step "named"`, `step "make"`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			in := filepath.Join(tmp, "ci.yaml")
			yml := "name: demo\non:\n  push: {}\njobs:\n  a:\n    runs-on: ubuntu-latest\n    steps:\n" + tc.steps

			if err := os.WriteFile(in, []byte(yml), 0o600); err != nil {
				t.Fatal(err)
			}

			hclDir := filepath.Join(tmp, "hcl")

			if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
				t.Fatalf("Unparse() error = %v", err)
			}

			got, err := os.ReadFile(filepath.Join(hclDir, "ci.hcl"))
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range tc.wantSteps {
				if !strings.Contains(string(got), want) {
					t.Errorf("HCL missing %q:\n%s", want, got)
				}
			}

			for _, unwanted := range tc.wantGone {
				if strings.Contains(string(got), unwanted) {
					t.Errorf("HCL should not hold %q:\n%s", unwanted, got)
				}
			}
		})
	}
}

// Dedup across jobs is the point of the registry and has to keep working: the
// same step in two jobs is one block both refer to.
func TestTheSameStepInTwoJobsIsOneBlock(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "ci.yaml")

	yml := "name: demo\non:\n  push: {}\njobs:\n" +
		"  a:\n    runs-on: ubuntu-latest\n    steps:\n      - id: cache\n        uses: actions/cache@v4\n      - run: make a\n" +
		"  b:\n    runs-on: ubuntu-latest\n    steps:\n      - id: cache\n        uses: actions/cache@v4\n      - run: make b\n"

	if err := os.WriteFile(in, []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	hclDir := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(hclDir, "ci.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	if n := strings.Count(string(got), `step "cache"`); n != 1 {
		t.Errorf("want 1 cache block, got %d:\n%s", n, got)
	}
}

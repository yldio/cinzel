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

// Two definitions naming the same file wrote one on top of the other. The
// output held whichever came last and nothing said the rest had gone.
func TestTwoDefinitionsCannotWriteToTheSameFile(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{name: "two workflows", hcl: twoWorkflowsHCL("shared", "shared")},
		{name: "two actions", hcl: twoActionsHCL("shared", "shared")},
		{
			// The same file on macOS and on Windows, so the same input would
			// otherwise produce different output depending on where it ran.
			name: "filenames differing only in case",
			hcl:  twoWorkflowsHCL("build", "Build"),
		},
		{
			name: "the same file reached by different paths",
			hcl:  twoWorkflowsHCL("sub/build", "sub/./build"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "in.hcl")

			if err := os.WriteFile(path, []byte(tc.hcl), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})

			if !errors.Is(err, errDuplicateFilename) {
				t.Fatalf("want errDuplicateFilename, got %v", err)
			}

			// The message has to name both, or there is nothing to go and fix.
			for _, want := range []string{"first", "second"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("want %q named in %q", want, err)
				}
			}
		})
	}
}

// Distinct filenames are the ordinary case and have to keep working, and a
// workflow and an action may share a name because they are written to
// different paths.
func TestDistinctFilenamesStillWork(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want []string
	}{
		{
			name: "two workflows",
			hcl:  twoWorkflowsHCL("one", "two"),
			want: []string{"one.yaml", "two.yaml"},
		},
		{
			name: "two actions",
			hcl:  twoActionsHCL("one", "two"),
			want: []string{filepath.Join("one", "action.yml"), filepath.Join("two", "action.yml")},
		},
		{
			name: "a workflow and an action sharing a name",
			hcl:  workflowAndActionHCL("same"),
			want: []string{"same.yaml", filepath.Join("same", "action.yml")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := parseWorkflow(t, tc.hcl)

			for _, want := range tc.want {
				if _, err := os.Stat(filepath.Join(out, want)); err != nil {
					t.Errorf("want %s written, got %v", want, err)
				}
			}
		})
	}
}

// workflowAndActionHCL returns one workflow and one action, both carrying
// filename.
func workflowAndActionHCL(filename string) string {
	return "step \"hi\" {\n" +
		"  run = \"echo hi\"\n" +
		"}\n\n" +
		"job \"b\" {\n" +
		"  runs_on {\n" +
		"    runners = \"ubuntu-latest\"\n" +
		"  }\n" +
		"  steps = [step.hi]\n" +
		"}\n\n" +
		workflowBlock("w", filename) +
		actionBlock("a", filename)
}

// twoWorkflowsHCL returns two workflows, "first" and "second", carrying the
// given filenames.
func twoWorkflowsHCL(first, second string) string {
	return "step \"hi\" {\n" +
		"  run = \"echo hi\"\n" +
		"}\n\n" +
		"job \"b\" {\n" +
		"  runs_on {\n" +
		"    runners = \"ubuntu-latest\"\n" +
		"  }\n" +
		"  steps = [step.hi]\n" +
		"}\n\n" +
		workflowBlock("first", first) +
		workflowBlock("second", second)
}

// twoActionsHCL returns two actions, "first" and "second", carrying the given
// filenames.
func twoActionsHCL(first, second string) string {
	return "step \"hi\" {\n" +
		"  run = \"echo hi\"\n" +
		"}\n\n" +
		actionBlock("first", first) +
		actionBlock("second", second)
}

func workflowBlock(id, filename string) string {
	return "workflow \"" + id + "\" {\n" +
		"  filename = \"" + filename + "\"\n" +
		"  name     = \"" + id + "\"\n" +
		"  on \"push\" {}\n" +
		"  jobs = [job.b]\n" +
		"}\n\n"
}

func actionBlock(id, filename string) string {
	return "action \"" + id + "\" {\n" +
		"  filename    = \"" + filename + "\"\n" +
		"  name        = \"" + id + "\"\n" +
		"  description = \"d\"\n" +
		"  runs {\n" +
		"    using = \"composite\"\n" +
		"    steps = [step.hi]\n" +
		"  }\n" +
		"}\n\n"
}

// A workflow written into a subdirectory and then renamed left the old file
// there, because the prune read only the top level of the output directory.
func TestRenamedNestedWorkflowIsPruned(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "in.hcl")
	output := filepath.Join(dir, "out")

	for _, filename := range []string{"sub/before", "sub/after"} {
		if err := os.WriteFile(path, []byte(workflowHCL(filename)), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: output}); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := os.Stat(filepath.Join(output, "sub", "before.yaml")); !os.IsNotExist(err) {
		t.Errorf("want the renamed-away file removed, stat err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(output, "sub", "after.yaml")); err != nil {
		t.Errorf("want the current file kept, got %v", err)
	}
}

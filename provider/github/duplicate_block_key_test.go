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

// Two blocks naming the same key wrote one on top of the other. The file held
// whichever came last and the command exited 0, so a workflow shipped with a
// value nobody in the file meant to set. The GitLab provider already refuses
// the same input for its variable blocks.
func TestTwoBlocksCannotWriteTheSameKey(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hcl   string
		named []string
	}{
		{
			name:  "two workflow env blocks",
			hcl:   workflowWithExtra("  env {\n    name = \"TOKEN\"\n    value = \"one\"\n  }\n\n  env {\n    name = \"TOKEN\"\n    value = \"two\"\n  }\n", ""),
			named: []string{"env", "TOKEN"},
		},
		{
			name:  "two workflow on blocks for one event",
			hcl:   workflowWithExtra("  on \"pull_request\" {\n    branches = [\"main\"]\n  }\n\n  on \"pull_request\" {\n    branches = [\"dev\"]\n  }\n", ""),
			named: []string{"on", "pull_request"},
		},
		{
			name:  "two job env blocks",
			hcl:   workflowWithExtra("", "  env {\n    name = \"J\"\n    value = \"one\"\n  }\n\n  env {\n    name = \"J\"\n    value = \"two\"\n  }\n"),
			named: []string{"env", "J"},
		},
		{
			name:  "two job output blocks",
			hcl:   workflowWithExtra("", "  output {\n    name = \"O\"\n    value = \"one\"\n  }\n\n  output {\n    name = \"O\"\n    value = \"two\"\n  }\n"),
			named: []string{"output", "O"},
		},
		{
			name:  "two job service blocks with one label",
			hcl:   workflowWithExtra("", "  service \"db\" {\n    image = \"postgres:one\"\n  }\n\n  service \"db\" {\n    image = \"postgres:two\"\n  }\n"),
			named: []string{"service", "db"},
		},
		{
			name:  "two env blocks nested in a container block",
			hcl:   workflowWithExtra("", "  container {\n    image = \"node\"\n\n    env {\n      name = \"C\"\n      value = \"one\"\n    }\n\n    env {\n      name = \"C\"\n      value = \"two\"\n    }\n  }\n"),
			named: []string{"env", "C"},
		},
		{
			name:  "two input blocks under one workflow_call trigger",
			hcl:   workflowWithExtra("  on \"workflow_call\" {\n    input \"k\" {\n      description = \"one\"\n    }\n\n    input \"k\" {\n      description = \"two\"\n    }\n  }\n", ""),
			named: []string{"input", "k"},
		},
		{
			name:  "two action input blocks",
			hcl:   actionWithExtra("  input \"k\" {\n    description = \"one\"\n  }\n\n  input \"k\" {\n    description = \"two\"\n  }\n"),
			named: []string{"input", "k"},
		},
		{
			name:  "two action output blocks",
			hcl:   actionWithExtra("  output \"o\" {\n    value = \"1\"\n  }\n\n  output \"o\" {\n    value = \"2\"\n  }\n"),
			named: []string{"output", "o"},
		},
		{
			name:  "two env blocks in an action's runs block",
			hcl:   actionRunsWithExtra("    env {\n      name = \"A\"\n      value = \"one\"\n    }\n\n    env {\n      name = \"A\"\n      value = \"two\"\n    }\n"),
			named: []string{"env", "A"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if !errors.Is(err, errDuplicateBlockKey) {
				t.Fatalf("want errDuplicateBlockKey, got %v", err)
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

// Distinct keys in the same position are what the check has to leave alone.
func TestBlocksWithDistinctKeysStillParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "two workflow env blocks",
			hcl:  workflowWithExtra("  env {\n    name = \"ONE\"\n    value = \"a\"\n  }\n\n  env {\n    name = \"TWO\"\n    value = \"b\"\n  }\n", ""),
		},
		{
			name: "two on blocks for different events",
			hcl:  workflowWithExtra("  on \"pull_request\" {\n  }\n\n  on \"workflow_dispatch\" {\n  }\n", ""),
		},
		{
			name: "two job service blocks with different labels",
			hcl:  workflowWithExtra("", "  service \"db\" {\n    image = \"postgres\"\n  }\n\n  service \"cache\" {\n    image = \"redis\"\n  }\n"),
		},
		{
			name: "two input blocks under one workflow_call trigger",
			hcl:  workflowWithExtra("  on \"workflow_call\" {\n    input \"one\" {\n      description = \"a\"\n    }\n\n    input \"two\" {\n      description = \"b\"\n    }\n  }\n", ""),
		},
		{
			name: "two action input blocks",
			hcl:  actionWithExtra("  input \"one\" {\n    description = \"a\"\n  }\n\n  input \"two\" {\n    description = \"b\"\n  }\n"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseHCLString(t, tc.hcl); err != nil {
				t.Fatalf("distinct keys must parse, got %v", err)
			}
		})
	}
}

// parseHCLString runs Parse over HCL written inline, returning its error.
func parseHCLString(t *testing.T, hcl string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "in.hcl")

	if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	return New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})
}

// workflowWithExtra returns a workflow carrying extra lines in its workflow
// block, its job block, or both.
func workflowWithExtra(workflowExtra, jobExtra string) string {
	return "step \"hi\" {\n  ignore_id = true\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n" +
		"  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		jobExtra +
		"  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n" +
		workflowExtra +
		"  jobs = [job.b]\n}\n"
}

// actionRunsWithExtra returns an action carrying extra lines inside its runs
// block, which is where an action's env blocks live.
func actionRunsWithExtra(extra string) string {
	return "action \"a\" {\n  filename = \"a\"\n  name = \"A\"\n  description = \"d\"\n\n" +
		"  runs {\n    using = \"node20\"\n    main = \"index.js\"\n\n" +
		extra +
		"  }\n}\n"
}

// actionWithExtra returns an action carrying extra lines in its action block.
func actionWithExtra(extra string) string {
	return "action \"a\" {\n  filename = \"a\"\n  name = \"A\"\n  description = \"d\"\n\n" +
		extra +
		"\n  runs {\n    using = \"node20\"\n    main = \"index.js\"\n  }\n}\n"
}

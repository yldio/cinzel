// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"strings"
	"testing"
)

// A block carrying no key of its own writes one whole YAML key, so a second
// one replaced the first and the run still exited 0. The permissions a
// workflow shipped with were whichever of the two the decoder happened to hand
// back last, which is not something the file says anywhere.
func TestASingletonBlockCannotBeWrittenTwice(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hcl   string
		named string
	}{
		{
			name:  "two workflow permissions blocks",
			hcl:   workflowWithExtra("  permissions {\n    contents = \"read\"\n  }\n\n  permissions {\n    contents = \"write\"\n  }\n", ""),
			named: "permissions",
		},
		{
			name:  "a workflow permissions attribute and block",
			hcl:   workflowWithExtra("  permissions = { contents = \"read\" }\n\n  permissions {\n    contents = \"write\"\n  }\n", ""),
			named: "permissions",
		},
		{
			name:  "two workflow defaults blocks",
			hcl:   workflowWithExtra("  defaults {\n    shell = \"bash\"\n  }\n\n  defaults {\n    shell = \"pwsh\"\n  }\n", ""),
			named: "defaults",
		},
		{
			name:  "two workflow concurrency blocks",
			hcl:   workflowWithExtra("  concurrency {\n    group = \"a\"\n  }\n\n  concurrency {\n    group = \"b\"\n  }\n", ""),
			named: "concurrency",
		},
		{
			name:  "two job runs_on blocks",
			hcl:   workflowWithExtra("", "  runs_on {\n    runners = \"macos-latest\"\n  }\n"),
			named: "runs_on",
		},
		{
			name:  "two job strategy blocks",
			hcl:   workflowWithExtra("", "  strategy {\n    fail_fast = true\n  }\n\n  strategy {\n    fail_fast = false\n  }\n"),
			named: "strategy",
		},
		{
			name:  "two matrix blocks in one strategy",
			hcl:   workflowWithExtra("", "  strategy {\n    matrix {\n      go = \"1.22\"\n    }\n\n    matrix {\n      go = \"1.23\"\n    }\n  }\n"),
			named: "matrix",
		},
		{
			name:  "two job container blocks",
			hcl:   workflowWithExtra("", "  container {\n    image = \"node\"\n  }\n\n  container {\n    image = \"go\"\n  }\n"),
			named: "container",
		},
		{
			name:  "a job container attribute and block",
			hcl:   workflowWithExtra("", "  container = { image = \"node\" }\n\n  container {\n    image = \"go\"\n  }\n"),
			named: "container",
		},
		{
			name:  "two job environment blocks",
			hcl:   workflowWithExtra("", "  environment {\n    name = \"x\"\n  }\n\n  environment {\n    name = \"y\"\n  }\n"),
			named: "environment",
		},
		{
			name:  "two job defaults blocks",
			hcl:   workflowWithExtra("", "  defaults {\n    shell = \"bash\"\n  }\n\n  defaults {\n    shell = \"pwsh\"\n  }\n"),
			named: "defaults",
		},
		{
			name:  "two job uses blocks",
			hcl:   "job \"b\" {\n  uses {\n    action = \"o/one/.github/workflows/x.yml\"\n    version = \"v1\"\n  }\n\n  uses {\n    action = \"o/two/.github/workflows/y.yml\"\n    version = \"v2\"\n  }\n}\n\nworkflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n",
			named: "uses",
		},
		{
			name:  "two action runs blocks",
			hcl:   "action \"a\" {\n  filename = \"a\"\n  name = \"A\"\n  description = \"d\"\n\n  runs {\n    using = \"node20\"\n    main = \"index.js\"\n  }\n\n  runs {\n    using = \"node20\"\n    main = \"other.js\"\n  }\n}\n",
			named: "runs",
		},
		{
			name:  "two action branding blocks",
			hcl:   actionWithExtra("  branding {\n    icon = \"check\"\n  }\n\n  branding {\n    icon = \"x\"\n  }\n"),
			named: "branding",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if !errors.Is(err, errDuplicateBlock) {
				t.Fatalf("want errDuplicateBlock, got %v", err)
			}

			// The message has to name the block, or there is nothing to go
			// and fix.
			if !strings.Contains(err.Error(), tc.named) {
				t.Errorf("want %q named in %q", tc.named, err)
			}
		})
	}
}

// One of each is what the check has to leave alone.
func TestASingletonBlockWrittenOnceStillParses(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "a workflow permissions block",
			hcl:  workflowWithExtra("  permissions {\n    contents = \"read\"\n  }\n", ""),
		},
		{
			name: "a workflow permissions attribute",
			hcl:  workflowWithExtra("  permissions = { contents = \"read\" }\n", ""),
		},
		{
			name: "a job strategy holding one matrix",
			hcl:  workflowWithExtra("", "  strategy {\n    fail_fast = true\n\n    matrix {\n      go = \"1.22\"\n    }\n  }\n"),
		},
		{
			name: "a job container block",
			hcl:  workflowWithExtra("", "  container {\n    image = \"node\"\n  }\n"),
		},
		{
			name: "an action runs block",
			hcl:  actionWithExtra("  branding {\n    icon = \"check\"\n  }\n"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseHCLString(t, tc.hcl); err != nil {
				t.Fatalf("one block must parse, got %v", err)
			}
		})
	}
}

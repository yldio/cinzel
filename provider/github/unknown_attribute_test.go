// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A step and a named block carry a "remain" body for the source range their
// comment is found at, and that body also collected every attribute nobody
// declared. A misspelling went out silently dropped and the run exited 0, while
// the same misspelling on a job or a workflow was an error. Worst of all,
// "uses" written as an attribute where the schema wants a block took the whole
// step with it and a bare "-" went out under "steps".
func TestAnUndeclaredAttributeIsReported(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hcl   string
		named []string
	}{
		{
			name:  "a stray attribute on a step",
			hcl:   stepWithExtra("  totalNonsense = \"xyz\"\n"),
			named: []string{"totalNonsense", "'s'"},
		},
		{
			name:  "a stray block in a step",
			hcl:   stepWithExtra("  bogus {\n    x = 1\n  }\n"),
			named: []string{"bogus", "'s'"},
		},
		{
			name: "uses written as an attribute on a step",
			// The schema wants a uses block, so the attribute form used to
			// leave a step with nothing in it at all.
			hcl:   stepWithExtra("  uses = \"actions/checkout@v4\"\n"),
			named: []string{"uses", "'s'"},
		},
		{
			name:  "a stray attribute in a job env block",
			hcl:   workflowWithExtra("", "  env {\n    name = \"K\"\n    value = \"v\"\n    totalNonsense = \"xyz\"\n  }\n"),
			named: []string{"totalNonsense"},
		},
		{
			name:  "a stray attribute in a workflow env block",
			hcl:   workflowWithExtra("  env {\n    name = \"K\"\n    value = \"v\"\n    totalNonsense = \"xyz\"\n  }\n", ""),
			named: []string{"totalNonsense"},
		},
		{
			name:  "a stray attribute in a job with block",
			hcl:   workflowWithExtra("", "  with {\n    name = \"K\"\n    value = \"v\"\n    totalNonsense = \"xyz\"\n  }\n"),
			named: []string{"totalNonsense"},
		},
		{
			name:  "a stray attribute in an action's env block",
			hcl:   actionRunsWithExtra("    env {\n      name = \"A\"\n      value = \"v\"\n      totalNonsense = \"xyz\"\n    }\n"),
			named: []string{"totalNonsense"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if err == nil {
				t.Fatal("want an error naming what was not expected, got nil")
			}

			// The message has to name what was written, or there is nothing to
			// go and fix.
			for _, want := range tc.named {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("want %q named in %q", want, err)
				}
			}
		})
	}
}

// What the schema does declare has to keep parsing, blocks included.
func TestDeclaredAttributesAndBlocksStillParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "a step with every block it declares",
			hcl: stepWithExtra("  uses {\n    action = \"actions/checkout\"\n    version = \"v4\"\n  }\n\n" +
				"  with {\n    name = \"w\"\n    value = \"1\"\n  }\n\n" +
				"  env {\n    name = \"E\"\n    value = \"v\"\n  }\n"),
		},
		{
			name: "a step with every attribute it declares",
			hcl: stepWithExtra("  name = \"n\"\n  if = \"true\"\n  shell = \"bash\"\n" +
				"  working_directory = \"/tmp\"\n  continue_on_error = true\n  timeout_minutes = 5\n"),
		},
		{
			name: "a named block with only name and value",
			hcl:  workflowWithExtra("", "  env {\n    name = \"K\"\n    value = \"v\"\n  }\n"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseHCLString(t, tc.hcl); err != nil {
				t.Fatalf("declared input must parse, got %v", err)
			}
		})
	}
}

// stepWithExtra returns a workflow whose step carries extra lines.
func stepWithExtra(extra string) string {
	return "step \"s\" {\n  ignore_id = true\n  run = \"echo a\"\n" +
		extra +
		"}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.s]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n"
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"strings"
	"testing"
)

// A step block setting nothing converted to an empty map, which went out as a
// bare "-" under "steps". GitHub rejects that file, and cinzel's own unparse
// reads the null node back and reports a step that must be an object, so a
// parse exited 0 on something no one could use and nothing could read back.
func TestAStepThatSetsNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "a job runs it",
			hcl: "step \"s\" {\n  ignore_id = true\n}\n\n" +
				"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.s]\n}\n\n" +
				"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n",
		},
		{
			name: "a job runs it after a step that sets something",
			hcl: "step \"ok\" {\n  ignore_id = true\n  run = \"echo a\"\n}\n\n" +
				"step \"s\" {\n  ignore_id = true\n}\n\n" +
				"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.ok, step.s]\n}\n\n" +
				"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n",
		},
		{
			name: "an action runs it",
			hcl: "step \"s\" {\n  ignore_id = true\n}\n\n" +
				"action \"a\" {\n  filename = \"a\"\n  name = \"A\"\n  description = \"d\"\n\n" +
				"  runs {\n    using = \"composite\"\n    steps = [step.s]\n  }\n}\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if !errors.Is(err, errEmptyStep) {
				t.Fatalf("want errEmptyStep, got %v", err)
			}

			// The message has to name the step, or there is nothing to go and
			// fix.
			if !strings.Contains(err.Error(), "'s'") {
				t.Errorf("want the step named in %q", err)
			}
		})
	}
}

// A step nothing runs is never emitted, and a step carrying anything at all is
// a step GitHub can run.
func TestAStepThatSetsSomethingStillParses(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "an empty step block nothing references",
			hcl: "step \"dead\" {\n  ignore_id = true\n}\n\n" +
				"step \"ok\" {\n  ignore_id = true\n  run = \"echo a\"\n}\n\n" +
				"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.ok]\n}\n\n" +
				"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n",
		},
		{
			name: "a step setting only a name",
			hcl:  stepWithExtra("  name = \"does nothing\"\n"),
		},
		{
			name: "a step keeping the id its label gives it",
			hcl: "step \"s\" {\n}\n\n" +
				"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.s]\n}\n\n" +
				"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseHCLString(t, tc.hcl); err != nil {
				t.Fatalf("want a parse, got %v", err)
			}
		})
	}
}

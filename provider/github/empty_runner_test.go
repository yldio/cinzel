// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"strings"
	"testing"
)

// runnerWorkflow returns a workflow whose only job carries the runs-on given.
func runnerWorkflow(runsOn string) string {
	return "name: w\non:\n  push:\njobs:\n  b:\n    runs-on: " + runsOn +
		"\n    steps:\n      - run: echo hi\n"
}

// A runs-on that is present but names no runner leaves GitHub with nothing to
// schedule the job on, and it refuses the workflow rather than running it:
// actionlint reports `"runs-on" section should not be empty` for the empty
// forms and `string should not be empty` for the empty names inside them. The
// check only asked whether the key was there, so every one of these went
// through, in both directions.
func TestARunsOnThatNamesNoRunnerIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "an empty string", yaml: runnerWorkflow(`""`)},
		{name: "an empty list", yaml: runnerWorkflow("[]")},
		{name: "a list holding an empty name", yaml: runnerWorkflow(`[""]`)},
		{name: "a list with one name empty", yaml: runnerWorkflow(`[ubuntu-latest, ""]`)},
		{name: "an empty group", yaml: runnerWorkflow("\n      group: \"\"")},
		{name: "an empty labels list", yaml: runnerWorkflow("\n      labels: []")},
		{name: "a labels list holding an empty name", yaml: runnerWorkflow("\n      labels: [\"\"]")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err == nil {
				t.Fatal("expected an error, got none")
			}

			if !errors.Is(err, errEmptyRunner) {
				t.Errorf("error is not the empty-runner one: %v", err)
			}
		})
	}
}

// The same runner on the way in, written as a runs_on block.
func TestARunsOnThatNamesNoRunnerIsRefusedFromHCL(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "an empty string", body: `    runners = ""`},
		{name: "an empty list", body: "    runners = []"},
		{name: "a list holding an empty name", body: `    runners = [""]`},
		{name: "an empty group", body: `    group = ""`},
		{name: "an empty labels list", body: "    labels = []"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, written := parseEmptyBlockHCL(t, runnerHCL(tc.body))

			if err == nil {
				t.Fatalf("expected an error, got:\n%s", written)
			}

			if !errors.Is(err, errEmptyRunner) {
				t.Errorf("error is not the empty-runner one: %v", err)
			}
		})
	}
}

// runnerHCL returns a workflow whose only job carries the runs_on body given.
func runnerHCL(body string) string {
	return "step \"s\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n" + body + "\n  }\n\n  steps = [step.s]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"
}

// A runner named, in any of the shapes GitHub reads it in, has to keep going
// through. An expression is left alone: GitHub resolves it at run time.
func TestARunnerThatIsNamedIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "a plain name", yaml: runnerWorkflow("ubuntu-latest")},
		{name: "a list of names", yaml: runnerWorkflow("[self-hosted, linux]")},
		{name: "a group", yaml: runnerWorkflow("\n      group: big")},
		{name: "a group and labels", yaml: runnerWorkflow("\n      group: big\n      labels: [gpu]")},
		{name: "an expression", yaml: runnerWorkflow("${{ matrix.os }}\n    strategy:\n      matrix:\n        os: [ubuntu-latest]")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err != nil {
				t.Fatalf("expected the runner to be kept, got %v", err)
			}
		})
	}

	err, written := parseEmptyBlockHCL(t, runnerHCL(`    runners = "ubuntu-latest"`))

	if err != nil {
		t.Fatalf("expected the runner to be kept, got %v", err)
	}

	if !strings.Contains(written, "runs-on: ubuntu-latest") {
		t.Errorf("lost the runner:\n%s", written)
	}
}

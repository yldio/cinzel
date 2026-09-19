// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"strings"
	"testing"
)

// emptyKeyWorkflow returns a workflow whose job body and step body carry
// whatever is passed, so a mapping can be dropped at either level.
func emptyKeyWorkflow(workflowExtra, jobExtra, stepExtra string) string {
	return "name: w\non:\n  push:\n" + workflowExtra +
		"jobs:\n  b:\n    runs-on: ubuntu-latest\n" + jobExtra +
		"    steps:\n      - run: echo hi\n" + stepExtra
}

// A YAML mapping key that is empty is one GitHub refuses outright
// ("string should not be empty"), and the HCL it unparsed to set a name
// attribute to "", which cinzel's own parse then refused to read back. So an
// unparse exited 0 on a file neither GitHub nor cinzel could use.
func TestAnEmptyMappingKeyIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "on a workflow env", yaml: emptyKeyWorkflow("env:\n  \"\": v\n", "", "")},
		{name: "on a job env", yaml: emptyKeyWorkflow("", "    env:\n      \"\": v\n", "")},
		{name: "on a job output", yaml: emptyKeyWorkflow("", "    outputs:\n      \"\": v\n", "")},
		{name: "on a step env", yaml: emptyKeyWorkflow("", "", "        env:\n          \"\": v\n")},
		{name: "on a step with", yaml: "name: w\non:\n  push:\njobs:\n  b:\n    runs-on: ubuntu-latest\n" +
			"    steps:\n      - uses: actions/checkout@v4\n        with:\n          \"\": v\n"},
		{name: "on a matrix axis", yaml: emptyKeyWorkflow("", "    strategy:\n      matrix:\n        \"\": [1]\n", "")},
		{name: "on a dispatch input", yaml: "name: w\non:\n  workflow_dispatch:\n    inputs:\n      \"\":\n        type: string\n" +
			"jobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err == nil {
				t.Fatal("expected an error, got none")
			}

			if !errors.Is(err, errEmptyKey) {
				t.Errorf("error is not the empty-key one: %v", err)
			}
		})
	}
}

// The same key on the way in, where a block label or a name attribute can be
// written empty just as easily.
func TestAnEmptyKeyIsRefusedFromHCL(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
	}{
		{name: "an on block label", src: workflowWithExtra("\n  on \"\" {}\n", "")},
		{name: "a service block label", src: workflowWithExtra("", "\n  service \"\" {\n    image = \"redis\"\n  }\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, written := parseEmptyBlockHCL(t, tc.src)

			if err == nil {
				t.Fatalf("expected an error, got:\n%s", written)
			}

			if !errors.Is(err, errEmptyKey) {
				t.Errorf("error is not the empty-key one: %v", err)
			}
		})
	}
}

// A key that says something has to keep going through, in both directions.
func TestAKeyThatSaysSomethingIsKept(t *testing.T) {
	err, dir := unparseYAMLString(t, emptyKeyWorkflow("env:\n  FOO: v\n", "    outputs:\n      out: v\n", ""))

	if err != nil {
		t.Fatalf("expected the workflow to be written, got %v", err)
	}

	if dir == "" {
		t.Fatal("no output directory")
	}

	parseErr, written := parseEmptyBlockHCL(t, workflowWithExtra("", "\n  service \"db\" {\n    image = \"redis\"\n  }\n"))

	if parseErr != nil {
		t.Fatalf("expected the service to be written, got %v", parseErr)
	}

	if !strings.Contains(written, "db:") {
		t.Errorf("lost the service:\n%s", written)
	}
}

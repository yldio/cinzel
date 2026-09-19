// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"strings"
	"testing"
)

// settingWorkflow returns a workflow carrying whatever is passed at the
// workflow, job and step level, so a setting can be emptied at any of them.
func settingWorkflow(workflowExtra, jobExtra, stepExtra string) string {
	return "name: w\non:\n  push:\n" + workflowExtra +
		"jobs:\n  b:\n    runs-on: ubuntu-latest\n" + jobExtra +
		"    steps:\n      - run: echo hi\n" + stepExtra
}

// A setting GitHub acts on and that says nothing leaves it nothing to act on,
// and it refuses the workflow rather than running it: actionlint reports
// `string should not be empty`. Both directions wrote one out and read it back
// without a word.
func TestASettingThatSaysNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "a workflow run-name", yaml: settingWorkflow("run-name: \"\"\n", "", "")},
		{name: "a workflow concurrency", yaml: settingWorkflow("concurrency: \"\"\n", "", "")},
		{name: "a workflow default shell", yaml: settingWorkflow("defaults:\n  run:\n    shell: \"\"\n", "", "")},
		{name: "a workflow default directory", yaml: settingWorkflow("defaults:\n  run:\n    working-directory: \"\"\n", "", "")},
		{name: "a job condition", yaml: settingWorkflow("", "    if: \"\"\n", "")},
		{name: "a job environment", yaml: settingWorkflow("", "    environment: \"\"\n", "")},
		{name: "a job container", yaml: settingWorkflow("", "    container: \"\"\n", "")},
		{name: "a job concurrency", yaml: settingWorkflow("", "    concurrency: \"\"\n", "")},
		{name: "a job default shell", yaml: settingWorkflow("", "    defaults:\n      run:\n        shell: \"\"\n", "")},
		{name: "a step script", yaml: settingWorkflow("", "", "      - run: \"\"\n")},
		{name: "a step shell", yaml: settingWorkflow("", "", "      - run: echo hi\n        shell: \"\"\n")},
		{name: "a step directory", yaml: settingWorkflow("", "", "      - run: echo hi\n        working-directory: \"\"\n")},
		{name: "a step condition", yaml: settingWorkflow("", "", "      - run: echo hi\n        if: \"\"\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err == nil {
				t.Fatal("a setting that says nothing was written to HCL")
			}

			if !errors.Is(err, errEmptyValue) {
				t.Fatalf("the error does not say the setting is the problem: %v", err)
			}
		})
	}
}

// A filter list GitHub matches on and that holds nothing, or holds an entry
// that says nothing, matches nothing at all. actionlint reports `"tags"
// section should not be empty` and `string should not be empty`.
func TestAFilterThatMatchesNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "an empty branches entry", yaml: "name: w\non:\n  push:\n    branches: [\"\"]\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
		{name: "an empty paths entry", yaml: "name: w\non:\n  push:\n    paths: [\"\"]\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
		{name: "an empty tags list", yaml: "name: w\non:\n  push:\n    tags: []\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
		{name: "an empty types entry", yaml: "name: w\non:\n  pull_request:\n    types: [\"\"]\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
		{name: "an empty matrix axis", yaml: settingWorkflow("", "    strategy:\n      matrix:\n        os: []\n", "")},
		{name: "an empty matrix axis value", yaml: settingWorkflow("", "    strategy:\n      matrix:\n        os: [\"\"]\n", "")},
		{name: "an empty needs list", yaml: settingWorkflow("", "    needs: []\n", "")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err == nil {
				t.Fatal("a filter that matches nothing was written to HCL")
			}

			if !errors.Is(err, errEmptyValue) {
				t.Fatalf("the error does not say the value is the problem: %v", err)
			}
		})
	}
}

// The same settings written in HCL reach the YAML through the other path, and
// were refused by neither.
func TestASettingThatSaysNothingIsRefusedFromHCL(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{name: "a step script", hcl: "step \"s\" {\n  run = \"\"\n}\n\njob \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.s]\n}\n\nworkflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"},
		{name: "a step shell", hcl: "step \"s\" {\n  run   = \"echo hi\"\n  shell = \"\"\n}\n\njob \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.s]\n}\n\nworkflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"},
		{name: "a job condition", hcl: "step \"s\" {\n  run = \"echo hi\"\n}\n\njob \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  if = \"\"\n\n  steps = [step.s]\n}\n\nworkflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"},
		{name: "an empty branches entry", hcl: "step \"s\" {\n  run = \"echo hi\"\n}\n\njob \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.s]\n}\n\nworkflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {\n    branches = [\"\"]\n  }\n\n  jobs = [job.b]\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if err == nil {
				t.Fatal("a setting that says nothing was written to YAML")
			}

			if !errors.Is(err, errEmptyValue) {
				t.Fatalf("the error does not say the setting is the problem: %v", err)
			}
		})
	}
}

// A name is the author's own text and GitHub takes an empty one, and so does
// an env, with or output value. None of these is a setting GitHub acts on.
func TestTextTheAuthorWroteIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "an empty workflow name", yaml: "name: \"\"\non:\n  push:\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
		{name: "an empty job name", yaml: settingWorkflow("", "    name: \"\"\n", "")},
		{name: "an empty step name", yaml: settingWorkflow("", "", "      - name: \"\"\n        run: echo hi\n")},
		{name: "an empty env value", yaml: settingWorkflow("env:\n  A: \"\"\n", "", "")},
		{name: "an empty output value", yaml: settingWorkflow("", "    outputs:\n      o: \"\"\n", "")},
		{name: "an empty with value", yaml: settingWorkflow("", "", "      - uses: actions/checkout@v4\n        with:\n          ref: \"\"\n")},
		{name: "a matrix axis holding a name", yaml: settingWorkflow("", "    strategy:\n      matrix:\n        os: [ubuntu-latest]\n", "")},
		{name: "a branches list holding a name", yaml: "name: w\non:\n  push:\n    branches: [main]\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"},
		{name: "a matrix axis carrying an expression", yaml: settingWorkflow("", "    strategy:\n      matrix:\n        os: ${{ fromJSON(needs.a.outputs.o) }}\n", "")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err != nil {
				t.Fatalf("text the author wrote was refused: %v", err)
			}
		})
	}
}

// A step keeps the shell it names, through both directions.
func TestASettingThatSaysSomethingIsKept(t *testing.T) {
	hcl := "step \"s\" {\n  run   = \"echo hi\"\n  shell = \"bash\"\n}\n\njob \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.s]\n}\n\nworkflow \"w\" {\n  filename = \"wf\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"

	got := parseHCLToYAML(t, hcl)

	if !strings.Contains(got, "shell: bash") {
		t.Fatalf("the shell the step named is gone:\n%s", got)
	}
}

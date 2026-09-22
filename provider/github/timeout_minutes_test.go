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

// timeoutWorkflow is a one-step workflow carrying jobAttr on the job and
// stepAttr on the step.
func timeoutWorkflow(stepAttr, jobAttr string) string {
	return `step "s" {
  ignore_id = true
  run       = "echo hi"
` + stepAttr + `}

job "b" {
  runs_on {
    runners = "ubuntu-latest"
  }
` + jobAttr + `
  steps = [step.s]
}

workflow "w" {
  filename = "w"

  on "push" {}

  jobs = [job.b]
}
`
}

// parseTimeoutHCL converts src and returns whatever the parse refused.
func parseTimeoutHCL(t *testing.T, src string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "w.hcl")

	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	return New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})
}

// GitHub gives a step or a job no time to run at all when the timeout is zero
// or less, and refuses the workflow rather than starting it: actionlint reads
// it as `value at "timeout-minutes" must be greater than zero`. A job's timeout
// was not checked at all, and a step's let zero through.
func TestATimeoutThatIsNotPositiveIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
	}{
		{name: "a step at zero", src: timeoutWorkflow("  timeout_minutes = 0\n", "")},
		{name: "a job at zero", src: timeoutWorkflow("", "  timeout_minutes = 0\n")},
		{name: "a job below zero", src: timeoutWorkflow("", "  timeout_minutes = -5\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseTimeoutHCL(t, tc.src)

			if err == nil {
				t.Fatal("expected the timeout to be refused")
			}

			if !strings.Contains(err.Error(), "greater than zero") {
				t.Errorf("error does not say the timeout must be positive: %v", err)
			}
		})
	}
}

// The same file read back from YAML has to be refused the same way, or a
// workflow GitHub will not start still converts into HCL.
func TestATimeoutThatIsNotPositiveIsRefusedFromYAML(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "a step at zero", yaml: "        timeout-minutes: 0\n"},
		{name: "a step below zero", yaml: "        timeout-minutes: -5\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, `name: w
on:
  push:
jobs:
  b:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`+tc.yaml)

			if err == nil {
				t.Fatal("expected the timeout to be refused")
			}

			if !strings.Contains(err.Error(), "greater than zero") {
				t.Errorf("error does not say the timeout must be positive: %v", err)
			}
		})
	}
}

// A positive timeout is the ordinary case, and an expression is resolved by
// GitHub at run time rather than by cinzel, so neither may be refused.
func TestATimeoutThatIsFineIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
	}{
		{name: "a step at one", src: timeoutWorkflow("  timeout_minutes = 1\n", "")},
		{name: "a job at thirty", src: timeoutWorkflow("", "  timeout_minutes = 30\n")},
		{name: "a job carrying an expression", src: timeoutWorkflow("", "  timeout_minutes = \"$${{ fromJSON(vars.T) }}\"\n")},
		{name: "neither setting one", src: timeoutWorkflow("", "")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseTimeoutHCL(t, tc.src); err != nil {
				t.Errorf("expected the timeout to be kept, got %v", err)
			}
		})
	}
}

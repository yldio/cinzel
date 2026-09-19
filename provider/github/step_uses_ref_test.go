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

// usesStepWorkflow is a one-step workflow whose uses block holds body.
func usesStepWorkflow(body string) string {
	return `step "s" {
  ignore_id = true

  uses {
` + body + `
  }
}

job "b" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.s]
}

workflow "w" {
  filename = "w"

  on "push" {}

  jobs = [job.b]
}
`
}

// parseUsesHCL converts src and returns whatever the parse refused.
func parseUsesHCL(t *testing.T, src string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "w.hcl")

	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	return New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})
}

// A step's uses reference was checked on the way back from YAML and not on the
// way out to it, so a parse wrote a reference GitHub cannot resolve and its own
// unparse then refused to read. actionlint reports all three as an invalid
// action format.
func TestAStepUsesReferenceIsChecked(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{
			name: "no owner",
			body: "    action  = \"\"\n    version = \"v4\"",
			want: "owner/repo@ref format",
		},
		{
			name: "no version after the at",
			body: "    action  = \"actions/checkout\"\n    version = \"\"",
			want: "empty version reference",
		},
		{
			name: "no version at all",
			body: "    action = \"actions/checkout\"",
			want: "version reference",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseUsesHCL(t, usesStepWorkflow(tc.body))

			if err == nil {
				t.Fatal("expected the reference to be refused")
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error does not say what is wrong with the reference: %v", err)
			}
		})
	}
}

// A local action carries no version and a docker one names an image, and both
// have to keep going through.
func TestAStepUsesReferenceThatIsFineIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "a remote action", body: "    action  = \"actions/checkout\"\n    version = \"v4\""},
		{name: "a local action", body: "    action = \"./.github/actions/setup\""},
		{name: "a parent-relative action", body: "    action = \"../setup\""},
		{name: "a docker action", body: "    action = \"docker://alpine:3.19\""},
		{name: "an action in a subdirectory", body: "    action  = \"actions/aws/ec2\"\n    version = \"main\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseUsesHCL(t, usesStepWorkflow(tc.body)); err != nil {
				t.Errorf("expected the reference to be kept, got %v", err)
			}
		})
	}
}

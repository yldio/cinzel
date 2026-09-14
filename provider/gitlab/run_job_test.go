// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A job written in the steps syntax carries its commands under "run" instead
// of "script". Parse used to reject one outright, so valid YAML failed.
func TestRunJobNeedsNoScript(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want []string
	}{
		{
			name: "run with an inline script",
			yaml: "job1:\n  run:\n    - name: build\n      script: make\n",
			want: []string{"run:", "name: build", "script: make"},
		},
		{
			name: "run with a step reference",
			yaml: "job1:\n  run:\n    - name: greet\n      step: ./steps/greet\n      inputs:\n        who: world\n",
			want: []string{"run:", "step: ./steps/greet", "who: world"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yaml)

			for _, want := range tc.want {
				if !strings.Contains(back, want) {
					t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", want, hcl, back)
				}
			}
		})
	}
}

// A job with neither "script" nor "run" is still a mistake worth reporting.
func TestJobWithoutScriptOrRunIsRejected(t *testing.T) {
	err := parseHCLString(t, "job \"job1\" {\n  stage = \"build\"\n}\n")

	if err == nil {
		t.Fatal("Parse() accepted a job with no script and no run")
	}

	if !strings.Contains(err.Error(), "must define 'script'") {
		t.Errorf("Parse() error = %v, want it to name the missing script", err)
	}
}

// A run job still needs a declared stage: "run" says where the commands are,
// not that the job inherits anything.
func TestRunJobStillNeedsItsStage(t *testing.T) {
	err := parseHCLString(t,
		"stages = [\"build\"]\n\njob \"job1\" {\n  run = [{ name = \"build\", script = \"make\" }]\n}\n")

	if err == nil {
		t.Fatal("Parse() accepted a run job with no stage")
	}

	if !strings.Contains(err.Error(), "must define 'stage'") {
		t.Errorf("Parse() error = %v, want it to name the missing stage", err)
	}
}

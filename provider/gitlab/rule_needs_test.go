// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A "needs" inside a rule names a job of this pipeline, the way the job-level
// one does, and GitLab refuses a pipeline whose rule waits on a job that is
// not in it. Only the job-level list was checked, so the YAML went out at exit
// 0 and GitLab rejected it on push.
func TestRuleNeedsMustNameADeclaredJob(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "plain name",
			hcl: `stages = ["build"]

job "a" {
  stage  = "build"
  script = ["make"]
  rule {
    if    = "$A"
    needs = ["ghost"]
  }
}
`,
		},
		{
			name: "object form",
			hcl: `stages = ["build"]

job "a" {
  stage  = "build"
  script = ["make"]
  rule {
    if = "$A"
    needs = [{
      job       = "ghost"
      artifacts = true
    }]
  }
}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)
			if err == nil {
				t.Fatal("want an error naming the unknown job")
			}

			if !strings.Contains(err.Error(), "ghost") {
				t.Fatalf("error does not name the job: %v", err)
			}
		})
	}
}

// The check has to stop at the name nothing declares. A rule waiting on a job
// the pipeline holds is left alone, in either form.
func TestRuleNeedsOnADeclaredJobIsKept(t *testing.T) {
	const yml = `stages:
  - build
b:
  stage: build
  script:
    - make b
a:
  stage: build
  script:
    - make
  rules:
    - if: "$A"
      needs:
        - b
        - job: b
          artifacts: true
`

	_, back := roundtripYAML(t, yml)

	if !strings.Contains(back, "needs:") {
		t.Fatalf("roundtrip lost the rule needs:\n%s", back)
	}
}

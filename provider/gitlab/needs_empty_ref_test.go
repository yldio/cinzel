// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A "needs" entry naming nothing was written as "job." with no identifier
// after the dot. That is not HCL, and nothing said so: the file was written
// and the command exited 0. Reading it back failed with "an attribute name is
// required after a dot", by which point the YAML it came from was gone.
func TestANeedsEntryNamingNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "empty string",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs:
    - ""
`,
		},
		{
			name: "empty job in an object",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs:
    - job: ""
      artifacts: true
`,
		},
		{
			name: "object naming neither a job nor a pipeline",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs:
    - {}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := unparseGitLab(t, tc.yaml)

			if err == nil {
				t.Fatalf("Unparse() error = nil, want an error\nHCL written:\n%s", got)
			}

			if !strings.Contains(err.Error(), "must name a job") {
				t.Errorf("Unparse() error = %q, want it to name the needs entry", err)
			}
		})
	}
}

// The refusal has to stop at the entry that names nothing. A need carrying a
// real job name comes back through parse, whichever form it takes: a local one
// as a reference to the job's label, so it follows a rename, and a remote one
// as the other pipeline's own name, which nothing here can rename.
func TestANeedNamingAJobStillRoundTrips(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yaml    string
		wantHCL string
	}{
		{
			name: "plain name",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs: [build]
`,
			wantHCL: "job.build,",
		},
		{
			name: "object form",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs:
    - job: build
      artifacts: true
`,
			wantHCL: "job       = job.build",
		},
		{
			name: "cross-project, no local job of that name",
			yaml: `test:
  script: [make test]
  needs:
    - job: build
      project: group/proj
      ref: main
`,
			wantHCL: `job     = "build"`,
		},
		{
			name: "cross-pipeline, which carries no project",
			yaml: `test:
  script: [make test]
  needs:
    - job: build
      pipeline: "$UPSTREAM_ID"
`,
			wantHCL: `job      = "build"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := unparseGitLab(t, tc.yaml)

			if err != nil {
				t.Fatalf("Unparse() error = %v", err)
			}

			if !strings.Contains(got, tc.wantHCL) {
				t.Errorf("HCL missing %q:\n%s", tc.wantHCL, got)
			}
		})
	}
}

// The reference a need carries is the job's own label, so it follows a job
// whose name HCL cannot spell. Sanitizing the name at the reference instead
// would still reach "build_app" here, which is why this asserts on the label
// the pipeline actually assigned: a second job sanitizing to the same thing
// takes "build_app_2", and a reference resolved from the name alone would
// point at the wrong one.
func TestANeedFollowsTheLabelTheJobWasGiven(t *testing.T) {
	got, err := unparseGitLab(t, `build-app:
  script: [make a]
build.app:
  script: [make b]
test:
  script: [make test]
  needs: ["build.app"]
`)
	if err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	if !strings.Contains(got, "job.build_app_2,") {
		t.Errorf("need did not follow the second job's label:\n%s", got)
	}
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A "needs" entry carrying "project" or "pipeline" names a job in another
// pipeline, which this file knows nothing about. Its name was written as a
// local job reference all the same, so it was sanitized like a local label and
// came back under a different name, and a name matching a local job followed
// that job's renames.
func TestCrossProjectNeedKeepsItsJobName(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
		want string
	}{
		{
			name: "name a local label could not carry",
			yml: `stages:
  - build
a:
  stage: build
  script:
    - make
  needs:
    - project: g/p
      job: "my job"
      ref: main
`,
			want: "job: my job",
		},
		{
			name: "name a local job also carries",
			yml: `stages:
  - build
"b-x":
  stage: build
  script:
    - make local
a:
  stage: build
  script:
    - make
  needs:
    - project: g/p
      job: "b-x"
      ref: main
`,
			want: "job: b-x",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, back := roundtripYAML(t, tc.yml)

			if !strings.Contains(back, tc.want) {
				t.Fatalf("roundtrip lost %q:\n%s", tc.want, back)
			}
		})
	}
}

// A file written before remote names were kept verbatim holds the name as a
// job reference. It still reads back, so the change does not strand HCL that
// is already on disk.
func TestCrossProjectNeedWrittenAsAReferenceStillReads(t *testing.T) {
	const hcl = `job "test" {
  script = ["make test"]
  need {
    job     = job.build
    project = "group/proj"
    ref     = "main"
  }
}
`

	if err := parseHCLString(t, hcl); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

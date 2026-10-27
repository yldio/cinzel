// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A job's "rules", "cache", "needs" and "services" each become a run of
// blocks, and they were written flush against each other and against the
// attributes around them: a wall of braces with no seam in it, unlike every
// other run of blocks these files hold.
func TestRepeatedJobBlocksAreSeparated(t *testing.T) {
	const pipeline = `build:
  stage: build
  script:
    - echo build
  rules:
    - if: '$A == "1"'
    - if: '$B == "2"'
  cache:
    - key: one
      paths: [a]
    - key: two
      paths: [b]
  services:
    - name: redis
    - name: postgres
`

	hcl := unparsePipeline(t, pipeline)

	for _, pair := range []string{"rule", "cache", "service"} {
		if strings.Contains(hcl, "}\n  "+pair+" {") {
			t.Errorf("two %q blocks are written flush:\n%s", pair, hcl)
		}
	}
}

// Several "include" entries are a run of blocks too.
func TestIncludeBlocksAreSeparated(t *testing.T) {
	hcl := unparsePipeline(t, "include:\n  - local: a.yml\n  - local: b.yml\nbuild:\n  script: [echo]\n")

	if strings.Contains(hcl, "}\ninclude {") {
		t.Errorf("two include blocks are written flush:\n%s", hcl)
	}
}

// The separating line goes above the comment, not between it and the block.
// Written the other way round, a note on a job's "services" was left stranded
// above the gap rather than sitting on what it describes.
func TestACommentStaysOnTheBlockItDescribes(t *testing.T) {
	const pipeline = `build:
  stage: build
  script: [echo]
  # above the services
  services:
    - name: redis
    - name: postgres
`

	hcl := unparsePipeline(t, pipeline)

	if !strings.Contains(hcl, "# above the services\n  service {") {
		t.Errorf("the comment is not on its block:\n%s", hcl)
	}
}

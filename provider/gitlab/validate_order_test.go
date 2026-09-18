// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// The job checks ran off a map range, so a pipeline with two jobs failing the
// same check reported whichever one the range reached first, and that changed
// from one run to the next: the same file gave a different error on a rerun,
// and a golden asserting the message failed at random.
func TestJobValidationReportsTheSameJobEveryRun(t *testing.T) {
	const hcl = `stages = ["build"]

job "alpha" {
  script = ["make a"]
}

job "beta" {
  script = ["make b"]
}
`

	for range 20 {
		err := parseHCLString(t, hcl)
		if err == nil {
			t.Fatal("want a missing stage error")
		}

		if !strings.Contains(err.Error(), "job 'alpha' must define 'stage'") {
			t.Fatalf("want the first job by name, got: %v", err)
		}
	}
}

// The keyword check ran off a map range too, so a pipeline with two jobs
// named after keywords reported whichever one the range reached first.
func TestKeywordNameReportsTheSameJobEveryRun(t *testing.T) {
	const hcl = `stages = ["build"]

job "image" {
  stage  = "build"
  script = ["make"]
}

job "stages" {
  stage  = "build"
  script = ["make"]
}
`

	for range 20 {
		err := parseHCLString(t, hcl)
		if err == nil {
			t.Fatal("want a keyword name error")
		}

		if !strings.Contains(err.Error(), "keyword: 'image'") {
			t.Fatalf("want the first job by name, got: %v", err)
		}
	}
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A leading dot is what marks a hidden key, which GitLab never runs and cinzel
// writes as a template block. A job block claiming one through its "id" was
// emitted as a hidden key all the same, so the job came back as a template and
// every reference to it broke on the way in.
func TestJobIDCannotBeHidden(t *testing.T) {
	const hcl = `stages = ["build"]

job "j" {
  id     = ".hidden"
  stage  = "build"
  script = ["make"]
}
`

	err := parseHCLString(t, hcl)
	if err == nil {
		t.Fatal("want an error naming the hidden id")
	}

	if !strings.Contains(err.Error(), ".hidden") {
		t.Fatalf("error does not name the id: %v", err)
	}
}

// A hidden job never runs, so nothing can wait on one. The reference was
// written as a job reference to a job that is not there, and unparse exited 0
// on a file the same tool's parse direction refused with "needs unknown job".
func TestNeedsOnAHiddenJobIsRefused(t *testing.T) {
	const yml = `stages: [build]
.hidden:
  stage: build
  script: [make]
k:
  stage: build
  script: [make]
  needs: [.hidden]
`

	err := unparseYAMLString(t, yml)
	if err == nil {
		t.Fatal("want an error naming the hidden job")
	}

	if !strings.Contains(err.Error(), ".hidden") {
		t.Fatalf("error does not name the job: %v", err)
	}
}

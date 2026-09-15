// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// countIn returns how many times sub occurs in s.
func countIn(s, sub string) int {
	return strings.Count(s, sub)
}

// A job may reference the same step twice, which the unparser encourages by
// collapsing identical steps onto one block. GitHub requires a step id to be
// unique within its job, so the repeat must not carry the id again.
func TestRepeatedStepEmitsOneID(t *testing.T) {
	got := parseHCLToYAML(t, `
step "build" {
  run = "make build"
}

step "check" {
  run = "make test"
}

job "a" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.build, step.check, step.build]
}

workflow "wf" {
  filename = "wf"

  on "push" {}

  jobs = [job.a]
}
`)

	if n := countIn(got, "id: build"); n != 1 {
		t.Errorf("want one 'id: build', got %d:\n%s", n, got)
	}

	if n := countIn(got, "run: make build"); n != 2 {
		t.Errorf("want the step run twice, got %d:\n%s", n, got)
	}
}

// The same id in two different jobs is fine: the uniqueness GitHub asks for is
// per job, not per workflow.
func TestSameStepInTwoJobsKeepsBothIDs(t *testing.T) {
	got := parseHCLToYAML(t, `
step "build" {
  run = "make build"
}

job "a" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.build]
}

job "b" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.build]
}

workflow "wf" {
  filename = "wf"

  on "push" {}

  jobs = [job.a, job.b]
}
`)

	if n := countIn(got, "id: build"); n != 2 {
		t.Errorf("want 'id: build' once per job, got %d:\n%s", n, got)
	}
}

// Steps referenced once each keep their ids, which is every other workflow.
func TestDistinctStepsKeepTheirIDs(t *testing.T) {
	got := parseHCLToYAML(t, `
step "one" {
  run = "echo one"
}

step "two" {
  run = "echo two"
}

job "a" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.one, step.two]
}

workflow "wf" {
  filename = "wf"

  on "push" {}

  jobs = [job.a]
}
`)

	for _, want := range []string{"id: one", "id: two"} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in output, got:\n%s", want, got)
		}
	}
}

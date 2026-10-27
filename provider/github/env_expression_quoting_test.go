// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// An expression is a plain scalar: the braces sit in the middle of it, where
// YAML reads them as ordinary characters. Quoting scanned the whole value for
// an indicator, so every "${{ ... }}" came back quoted and a generated file
// no longer matched the hand-written one it replaced.
func TestAnExpressionIsNotQuoted(t *testing.T) {
	got := parseHCLToYAML(t, `
workflow "wf" {
  filename = "wf"

  name = "ci"

  on "push" {}

  env {
    name  = "CMS_IMAGE_TAG"
    value = "$${{ github.sha }}"
  }

  jobs = [job.build]
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

step "echo" {
  ignore_id = true

  run = "echo hi"
}
`)

	if !strings.Contains(got, "CMS_IMAGE_TAG: ${{ github.sha }}") {
		t.Errorf("the expression came back quoted:\n%s", got)
	}
}

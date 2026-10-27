// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A job's "services" becomes one block per service, and they were written
// flush against each other. Every other run of blocks in these files has a
// blank line between them.
func TestServiceBlocksAreSeparated(t *testing.T) {
	const workflow = `name: ci
on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis
      postgres:
        image: postgres
    steps:
      - run: make
`

	hcl := unparse(t, workflow)

	if strings.Contains(hcl, "}\n  service \"") {
		t.Errorf("two service blocks are written flush:\n%s", hcl)
	}
}

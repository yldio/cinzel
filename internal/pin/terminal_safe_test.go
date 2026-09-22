// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// A warning quotes the action and version it read out of the file, and either
// is free to carry an ANSI escape sequence. Written to a terminal as they
// stand, those sequences are acted on rather than shown: a crafted action name
// can erase the warning that names it and print a summary of its own in its
// place. The errors this tool ends on are already escaped for the same reason.
func TestAWarningDoesNotCarryAnEscapeSequenceToTheTerminal(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "acme/x\u001b[2K\rPin summary: 9 pinned"
    version = "main"
  }
}

step "b" {
  uses {
    action  = "acme/real"
    version = "v4\u001b[31m"
  }
}
`)

	resolver := &mockResolver{shas: map[string]string{}}

	var buf bytes.Buffer

	if _, err := PinFile(context.Background(), path, resolver, &buf, false); err != nil {
		t.Fatal(err)
	}

	if strings.ContainsAny(buf.String(), "\x1b\r") {
		t.Errorf("a control character reached the output: %q", buf.String())
	}
}

// The same on the upgrade side.
func TestAnUpgradeWarningDoesNotCarryAnEscapeSequence(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "acme/x\u001b[2K\rUpgrade summary: 9 upgraded"
    version = "v1"
  }
}
`)

	resolver := &upgraderStub{}

	var buf bytes.Buffer

	if _, err := UpgradeFile(context.Background(), path, resolver, &buf, false); err != nil {
		t.Fatal(err)
	}

	if strings.ContainsAny(buf.String(), "\x1b\r") {
		t.Errorf("a control character reached the output: %q", buf.String())
	}
}

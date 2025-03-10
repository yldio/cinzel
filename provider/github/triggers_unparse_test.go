// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"testing"

	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
)

func TestTriggerBlockTypeForEventKey(t *testing.T) {
	tests := []struct {
		event    string
		key      string
		expect   string
		expectOK bool
	}{
		{event: "workflow_call", key: "inputs", expect: "input", expectOK: true},
		{event: "workflow_call", key: "outputs", expect: "output", expectOK: true},
		{event: "workflow_dispatch", key: "inputs", expect: "input", expectOK: true},
		{event: "push", key: "inputs", expect: "", expectOK: false},
	}

	for _, tt := range tests {
		got, ok := ghworkflow.TriggerBlockTypeForEventKey(tt.event, tt.key)
		if ok != tt.expectOK {
			t.Fatalf("expected ok=%v, got %v", tt.expectOK, ok)
		}

		if got != tt.expect {
			t.Fatalf("expected block type %q, got %q", tt.expect, got)
		}
	}
}

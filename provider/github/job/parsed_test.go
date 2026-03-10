// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package job

import "testing"

func TestNewParsed(t *testing.T) {
	body := map[string]any{
		"stepsRefs": []string{"step_build", "step_test"},
		"runs-on":   "ubuntu-latest",
	}

	p := NewParsed("build", body)

	if p.ID != "build" {
		t.Fatalf("expected build, got %s", p.ID)
	}

	if len(p.StepRefs) != 2 {
		t.Fatalf("expected 2 step refs, got %d", len(p.StepRefs))
	}

	if _, ok := p.Body["stepsRefs"]; ok {
		t.Fatal("expected stepsRefs to be removed from body")
	}

	if _, ok := p.Body["runs-on"]; !ok {
		t.Fatal("expected runs-on to remain in body")
	}
}

func TestNewParsedNoStepRefs(t *testing.T) {
	body := map[string]any{
		"runs-on": "ubuntu-latest",
	}

	p := NewParsed("deploy", body)

	if len(p.StepRefs) != 0 {
		t.Fatalf("expected 0 step refs, got %d", len(p.StepRefs))
	}
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import "testing"

func TestNormalizeOnEvent(t *testing.T) {
	raw := map[string]any{
		"input":  map[string]any{"a": map[string]any{"type": "string"}},
		"output": map[string]any{"o": map[string]any{"value": "1"}},
		"secret": map[string]any{"s": map[string]any{"required": true}},
	}

	out := NormalizeOnEvent("workflow_call", raw)

	if _, ok := out["inputs"]; !ok {
		t.Fatalf("expected inputs key after normalization, got %#v", out)
	}
	if _, ok := out["outputs"]; !ok {
		t.Fatalf("expected outputs key after normalization, got %#v", out)
	}
	if _, ok := out["secrets"]; !ok {
		t.Fatalf("expected secrets key after normalization, got %#v", out)
	}
}

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
		got, ok := TriggerBlockTypeForEventKey(tt.event, tt.key)
		if ok != tt.expectOK {
			t.Fatalf("expected ok=%v, got %v", tt.expectOK, ok)
		}
		if got != tt.expect {
			t.Fatalf("expected block type %q, got %q", tt.expect, got)
		}
	}
}

func TestValidateModel(t *testing.T) {
	err := ValidateModel(ValidationModel{HasOn: true, OnCount: 1, JobRefs: []string{"build"}})
	if err != nil {
		t.Fatalf("expected valid workflow model, got %v", err)
	}

	err = ValidateModel(ValidationModel{HasOn: false, OnCount: 0, JobRefs: []string{"build"}})
	if err == nil {
		t.Fatal("expected error for missing on")
	}
}

func TestNewYAMLDocument(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	doc, ok, err := NewYAMLDocument(map[string]any{"on": map[string]any{"push": map[string]any{}}, "jobs": map[string]any{"build": map[string]any{}}}, mapper)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ok {
		t.Fatal("expected workflow document")
	}

	if doc.On == nil || doc.Jobs == nil {
		t.Fatalf("expected on/jobs to be present, got %#v", doc)
	}
}

func TestNormalizeOnShorthandString(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	on, err := NormalizeOn("push", mapper)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := on["push"]; !ok {
		t.Fatalf("expected push event, got %#v", on)
	}
}

func TestNormalizeOnShorthandList(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	on, err := NormalizeOn([]any{"push", "pull_request"}, mapper)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := on["push"]; !ok {
		t.Fatalf("expected push event, got %#v", on)
	}

	if _, ok := on["pull_request"]; !ok {
		t.Fatalf("expected pull_request event, got %#v", on)
	}
}

func TestNormalizeOnSchedule(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	on, err := NormalizeOn(map[string]any{
		"schedule": []any{
			map[string]any{"cron": "0 0 * * *"},
			map[string]any{"cron": "0 12 * * 1"},
		},
	}, mapper)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	schedule, ok := on["schedule"].(map[string]any)
	if !ok {
		t.Fatalf("expected schedule object, got %#v", on["schedule"])
	}

	cron, ok := schedule["cron"].([]any)
	if !ok || len(cron) != 2 {
		t.Fatalf("expected two cron values, got %#v", schedule["cron"])
	}
}

func TestValidateModelDuplicateJobRefs(t *testing.T) {
	err := ValidateModel(ValidationModel{HasOn: true, OnCount: 1, JobRefs: []string{"build", "build"}})
	if err == nil {
		t.Fatal("expected error for duplicate job refs")
	}
}

func TestValidateModelNoJobs(t *testing.T) {
	err := ValidateModel(ValidationModel{HasOn: true, OnCount: 1, JobRefs: []string{}})
	if err == nil {
		t.Fatal("expected error for no jobs")
	}
}

func TestNewParsed(t *testing.T) {
	body := map[string]any{
		"filename": "ci",
		"jobsRefs": []string{"build", "test"},
		"on":       map[string]any{"push": map[string]any{}},
	}

	p := NewParsed("my_workflow", body)
	if p.Filename != "ci" {
		t.Fatalf("expected ci, got %s", p.Filename)
	}
	if len(p.JobRefs) != 2 {
		t.Fatalf("expected 2 job refs, got %d", len(p.JobRefs))
	}
	if _, ok := p.Body["filename"]; ok {
		t.Fatal("expected filename to be removed from body")
	}
	if _, ok := p.Body["jobsRefs"]; ok {
		t.Fatal("expected jobsRefs to be removed from body")
	}
	if _, ok := p.Body["on"]; !ok {
		t.Fatal("expected on to remain in body")
	}
}

func TestDenormalizeScheduleEvent(t *testing.T) {
	t.Run("multiple crons", func(t *testing.T) {
		normalized := map[string]any{
			"cron": []any{"0 0 * * *", "0 12 * * 1"},
		}
		items := DenormalizeScheduleEvent(normalized)
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("single cron string", func(t *testing.T) {
		normalized := map[string]any{
			"cron": "0 0 * * *",
		}
		items := DenormalizeScheduleEvent(normalized)
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
	})

	t.Run("no cron key", func(t *testing.T) {
		normalized := map[string]any{"other": "val"}
		items := DenormalizeScheduleEvent(normalized)
		if len(items) != 1 {
			t.Fatalf("expected 1 item (passthrough), got %d", len(items))
		}
	})
}

func TestNewYAMLDocumentNoWorkflow(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	_, ok, err := NewYAMLDocument(map[string]any{"name": "test"}, mapper)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected not a workflow document")
	}
}

func TestNormalizeOnEmptyString(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	_, err := NormalizeOn("", mapper)
	if err == nil {
		t.Fatal("expected error for empty string event")
	}
}

func TestNormalizeOnEmptyListEntry(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	_, err := NormalizeOn([]any{"push", ""}, mapper)
	if err == nil {
		t.Fatal("expected error for empty list entry")
	}
}

func TestNormalizeOnBoolEvent(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	on, err := NormalizeOn(map[string]any{"push": true}, mapper)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := on["push"]; !ok {
		t.Fatalf("expected push event, got %#v", on)
	}
}

func TestNormalizeOnInvalidShape(t *testing.T) {
	mapper := func(v any) (map[string]any, bool) {
		m, ok := v.(map[string]any)
		return m, ok
	}

	_, err := NormalizeOn(123, mapper)
	if err == nil {
		t.Fatal("expected error for invalid on shape")
	}
}

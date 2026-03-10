// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamlwriter

import (
	"testing"
)

type mockUpdater struct {
	filename string
	Content  string `yaml:"content"`
}

func (m mockUpdater) GetFilename() string         { return m.filename }
func (m mockUpdater) Validation() error           { return nil }
func (m mockUpdater) PostChanges(b []byte) []byte { return b }

func TestWriterDo(t *testing.T) {
	w := New([]mockUpdater{
		{filename: "test", Content: "hello"},
	})

	result, err := w.Do()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := result["test.yaml"]; !ok {
		t.Fatalf("expected test.yaml key, got %v", result)
	}

	content := string(result["test.yaml"])

	if content == "" {
		t.Fatal("expected non-empty YAML content")
	}
}

func TestWriterDoMultiple(t *testing.T) {
	w := New([]mockUpdater{
		{filename: "first", Content: "a"},
		{filename: "second", Content: "b"},
	})

	result, err := w.Do()
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result))
	}

	if _, ok := result["first.yaml"]; !ok {
		t.Fatal("expected first.yaml")
	}

	if _, ok := result["second.yaml"]; !ok {
		t.Fatal("expected second.yaml")
	}
}

type failingUpdater struct {
	mockUpdater
}

func (f failingUpdater) Validation() error { return errValidation }

var errValidation = &validationError{}

type validationError struct{}

func (e *validationError) Error() string { return "validation failed" }

func TestWriterDoValidationError(t *testing.T) {
	w := New([]failingUpdater{
		{mockUpdater: mockUpdater{filename: "test", Content: "hello"}},
	})

	_, err := w.Do()

	if err == nil {
		t.Fatal("expected validation error")
	}
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"testing"
)

func TestSplitYAMLDocuments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "single document",
			input: "name: test\non:\n  push:",
			want:  1,
		},
		{
			name:  "two documents",
			input: "name: workflow1\non:\n  push:\n---\nname: workflow2\non:\n  pull_request:",
			want:  2,
		},
		{
			name:  "leading separator ignored",
			input: "---\nname: test",
			want:  1,
		},
		{
			name:  "three documents",
			input: "name: a\n---\nname: b\n---\nname: c",
			want:  3,
		},
		{
			name:  "empty input",
			input: "",
			want:  0,
		},
		{
			name:  "whitespace only",
			input: "   \n\n  ",
			want:  0,
		},
		{
			name:  "separator with whitespace",
			input: "name: a\n  ---  \nname: b",
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitYAMLDocuments(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitYAMLDocuments() returned %d documents, want %d\ndocs: %v", len(got), tt.want, got)
			}
		})
	}
}

func TestSplitYAMLDocumentsContent(t *testing.T) {
	input := "name: workflow1\non:\n  push:\n---\nname: workflow2\non:\n  pull_request:"
	docs := splitYAMLDocuments(input)

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	if got := docs[0]; got != "name: workflow1\non:\n  push:\n" {
		t.Errorf("doc[0]:\ngot:  %q\nwant: %q", got, "name: workflow1\non:\n  push:\n")
	}

	if got := docs[1]; got != "name: workflow2\non:\n  pull_request:\n" {
		t.Errorf("doc[1]:\ngot:  %q\nwant: %q", got, "name: workflow2\non:\n  pull_request:\n")
	}
}

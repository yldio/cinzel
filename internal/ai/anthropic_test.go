// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"testing"
)

func TestStripFences(t *testing.T) {
	tests := []struct {
		name string
		input string
		want  string
	}{
		{
			name:  "no fences",
			input: "name: test\non:\n  push:",
			want:  "name: test\non:\n  push:",
		},
		{
			name:  "yaml fence",
			input: "```yaml\nname: test\non:\n  push:\n```",
			want:  "name: test\non:\n  push:",
		},
		{
			name:  "yml fence",
			input: "```yml\nname: test\n```",
			want:  "name: test",
		},
		{
			name:  "bare fence",
			input: "```\nname: test\n```",
			want:  "name: test",
		},
		{
			name:  "fence with surrounding text",
			input: "Here is the workflow:\n\n```yaml\nname: test\n```\n\nHope this helps!",
			want:  "name: test",
		},
		{
			name:  "multiple fences joined with separator",
			input: "```yaml\nname: workflow1\n```\n\n```yaml\nname: workflow2\n```",
			want:  "name: workflow1\n---\nname: workflow2",
		},
		{
			name:  "whitespace only input",
			input: "   \n\n  ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripFences(tt.input)
			if got != tt.want {
				t.Errorf("StripFences():\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

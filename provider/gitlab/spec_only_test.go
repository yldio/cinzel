// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// The root document was appended whatever it held, so a pipeline that is
// nothing but a spec header wrote "spec: ..." and then "---" and "{}".
func TestSpecOnlyPipelineEmitsOneDocument(t *testing.T) {
	spec := map[string]any{"inputs": map[string]any{"env": map[string]any{"default": "staging"}}}

	t.Run("a spec on its own", func(t *testing.T) {
		out, err := marshalPipelineYAML(map[string]any{specKey: spec}, nil)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got := string(out)

		if strings.Contains(got, "---") {
			t.Errorf("want one document, got:\n%s", got)
		}

		if strings.Contains(got, "{}") {
			t.Errorf("want no empty root, got:\n%s", got)
		}
	})

	t.Run("a spec with jobs still writes both", func(t *testing.T) {
		out, err := marshalPipelineYAML(map[string]any{
			specKey: spec,
			"build": map[string]any{"script": []any{"make"}},
		}, nil)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got := string(out)

		if !strings.Contains(got, "---") {
			t.Errorf("want two documents, got:\n%s", got)
		}

		if !strings.Contains(got, "build:") {
			t.Errorf("want the job in the second document, got:\n%s", got)
		}
	})

	// Nothing at all still has to write a document a reader accepts.
	t.Run("an empty pipeline keeps its root", func(t *testing.T) {
		out, err := marshalPipelineYAML(map[string]any{}, nil)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		if got := strings.TrimSpace(string(out)); got != "{}" {
			t.Errorf("want {}, got %q", got)
		}
	})
}

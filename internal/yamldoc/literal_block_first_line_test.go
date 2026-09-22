// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc_test

import (
	"testing"

	"github.com/yldio/cinzel/internal/yamldoc"
	yamlv3 "gopkg.in/yaml.v3"
)

// TestAValueAnIndentedBlockCannotStateIsQuotedInstead covers the strings whose
// first line a literal block has no way to write. A block indents its content
// and reads the indentation off the first line, so a first line that starts
// with a tab produces a file no reader accepts, and an empty one puts its blank
// lines above the indicator that would describe them.
func TestAValueAnIndentedBlockCannotStateIsQuotedInstead(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"a script indented with tabs":     "\techo one\n\techo two",
		"a tab on a line of its own":      "\t\necho two",
		"a leading blank line":            "\n\necho two",
		"nothing but newlines":            "\n\n",
		"a blank line above a tabbed one": "\n\techo two",
	}

	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc := yamldoc.New()
			doc.Set("run", yamldoc.Scalar(run))

			encoded, err := yamldoc.Encode(doc)
			if err != nil {
				t.Fatalf("encoding %q failed: %v", run, err)
			}

			var back map[string]any
			if err := yamlv3.Unmarshal(encoded, &back); err != nil {
				t.Fatalf("the YAML written for %q cannot be read back: %v\nwire:\n%s", run, err, encoded)
			}

			got, ok := back["run"].(string)
			if !ok {
				t.Fatalf("run came back as %T, not a string\nwire:\n%s", back["run"], encoded)
			}

			if got != run {
				t.Errorf("run came back changed:\n want %q\n  got %q\nwire:\n%s", run, got, encoded)
			}
		})
	}
}

// TestAValueAnIndentedBlockCanStateKeepsIt is the control. The fix narrows
// which strings get a literal block, so a run script that reads perfectly well
// as one has to keep it: a test that only checks the quoted cases would pass
// against an encoder that quoted everything.
func TestAValueAnIndentedBlockCanStateKeepsIt(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"an ordinary script":          "echo one\necho two",
		"a tab below the first line":  "echo one\n\techo two",
		"a space-indented first line": " echo one\necho two",
		"trailing blank lines":        "echo one\n\n\n",
	}

	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc := yamldoc.New()
			doc.Set("run", yamldoc.Scalar(run))

			encoded, err := yamldoc.Encode(doc)
			if err != nil {
				t.Fatalf("encoding %q failed: %v", run, err)
			}

			var node yamlv3.Node
			if err := yamlv3.Unmarshal(encoded, &node); err != nil {
				t.Fatalf("the YAML written for %q cannot be read back: %v\nwire:\n%s", run, err, encoded)
			}

			value := node.Content[0].Content[1]
			if value.Style != yamlv3.LiteralStyle {
				t.Errorf("%q lost its literal block\nwire:\n%s", run, encoded)
			}

			if value.Value != run {
				t.Errorf("run came back changed:\n want %q\n  got %q\nwire:\n%s", run, value.Value, encoded)
			}
		})
	}
}

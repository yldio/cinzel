// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"testing"
)

// matrixHCL wraps a matrix body in the smallest workflow that carries it.
func matrixHCL(matrixBody string) string {
	return workflowWithExtra("", "  strategy {\n    matrix {\n"+matrixBody+"    }\n  }\n")
}

// GitHub spreads a job over the values an axis lists, and reads "include" and
// "exclude" as lists of whole combinations. HCL writes a mapping as a block, so
// each of these was written as a block and read as a mapping where a list
// belongs. GitHub refuses every one of them, and actionlint reports `"matrix
// values" section must be sequence node but got mapping node`, `"include"
// section must be sequence node but got mapping node` and `element in "include"
// section is sequence node but mapping node is expected`.
func TestAMatrixThatSpreadsOverNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want error
	}{
		{
			name: "an axis written as a block",
			hcl:  matrixHCL("      go_version {\n        a = \"1\"\n      }\n"),
			want: errMatrixAxisShape,
		},
		{
			name: "an axis written as a mapping attribute",
			hcl:  matrixHCL("      go_version = { a = \"1\" }\n"),
			want: errMatrixAxisShape,
		},
		{
			name: "an include written as a block",
			hcl:  matrixHCL("      go = [\"1\"]\n\n      include {\n        go = \"1\"\n      }\n"),
			want: errMatrixEntryShape,
		},
		{
			name: "an exclude written as a block",
			hcl:  matrixHCL("      go = [\"1\"]\n\n      exclude {\n        go = \"1\"\n      }\n"),
			want: errMatrixEntryShape,
		},
		{
			name: "an include entry that is not a combination",
			hcl:  matrixHCL("      go = [\"1\"]\n\n      include = [\"x\"]\n"),
			want: errMatrixEntryShape,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if err == nil {
				t.Fatal("expected an error, got none")
			}

			if !errors.Is(err, tc.want) {
				t.Errorf("error is not the matrix shape one: %v", err)
			}
		})
	}
}

// The same shapes read from YAML, which is where a hand-written workflow
// carrying one arrives.
func TestAMatrixThatSpreadsOverNothingIsRefusedFromYAML(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want error
	}{
		{
			name: "an axis holding a mapping",
			yaml: "        go_version:\n          a: \"1\"\n",
			want: errMatrixAxisShape,
		},
		{
			name: "an include holding a mapping",
			yaml: "        go: [\"1\"]\n        include:\n          go: \"1\"\n",
			want: errMatrixEntryShape,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yml := "name: w\non:\n  push:\njobs:\n  b:\n    runs-on: ubuntu-latest\n" +
				"    strategy:\n      matrix:\n" + tc.yaml +
				"    steps:\n      - run: echo hi\n"

			err, written := unparseYAMLString(t, yml)

			if err == nil {
				t.Fatalf("expected an error, got:\n%s", written)
			}

			if !errors.Is(err, tc.want) {
				t.Errorf("error is not the matrix shape one: %v", err)
			}
		})
	}
}

// What GitHub does spread over has to keep going through: a list, an
// expression that produces one, and an include listing whole combinations.
func TestAMatrixThatSpreadsOverSomethingIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "an axis listing its values",
			hcl:  matrixHCL("      go = [\"1.21\", \"1.22\"]\n"),
		},
		{
			name: "an axis written as a variable block",
			hcl:  matrixHCL("      variable {\n        name  = \"go\"\n        value = [\"1.21\"]\n      }\n"),
		},
		{
			name: "an axis carrying an expression",
			hcl:  matrixHCL("      go = \"$${{ fromJSON(needs.a.outputs.v) }}\"\n"),
		},
		{
			name: "an include listing combinations",
			hcl:  matrixHCL("      go = [\"1.21\"]\n\n      include = [{\n        go = \"1.21\"\n      }]\n"),
		},
		{
			name: "an empty include, which adds nothing",
			hcl:  matrixHCL("      go = [\"1.21\"]\n\n      include = []\n"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := parseHCLString(t, tc.hcl); err != nil {
				t.Fatalf("a matrix GitHub spreads over must parse, got %v", err)
			}
		})
	}
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// The unparse path decodes with yaml.v3 while the rest of the repository
// decodes with goccy, and the two disagree on how a scalar resolves. Nothing
// in the golden fixtures carries a value that tells the difference, so the
// cases where a value could come back as another value are pinned here.
func TestScalarsSurviveTheDecodeBoundary(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{
			// yaml.v3 resolves this to a float, which cannot hold twenty
			// digits: it used to be emitted as 100000000000000000000.
			name:  "whole number too large for an integer",
			value: "99999999999999999999",
			want:  `value = "99999999999999999999"`,
		},
		{
			name:  "negative whole number too large for an integer",
			value: "-99999999999999999999",
			want:  `value = "-99999999999999999999"`,
		},
		{
			name:  "largest integer that still fits",
			value: "9223372036854775807",
			want:  "value = 9223372036854775807",
		},
		{
			// A float is meant to stay a float, so the retagging must not
			// reach it.
			name:  "float keeps its type",
			value: "1.5",
			want:  "value = 1.5",
		},
		{
			name:  "exponent keeps its type",
			value: "1e5",
			want:  "value = 100000",
		},
		{
			// YAML 1.1 read this as sixty seconds past one. YAML 1.2 does
			// not, and neither should the output.
			name:  "sexagesimal is a string",
			value: "1:30",
			want:  `value = "1:30"`,
		},
		{
			name:  "off is a string, not a boolean",
			value: "off",
			want:  `value = "off"`,
		},
		{
			name:  "NO is a string, not a boolean",
			value: "NO",
			want:  `value = "NO"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := unparseEnvValue(t, tc.value)

			if !strings.Contains(got, tc.want) {
				t.Errorf("want %s in the output, got:\n%s", tc.want, got)
			}
		})
	}
}

// An anchor, an alias and a merge key are resolved by the decoder, so what
// reaches the HCL writer is the expanded document. The expansion is what the
// output has to reflect.
func TestAnchorsAndMergeKeysExpand(t *testing.T) {
	yaml := "on: push\n" +
		"jobs:\n" +
		"  base:\n" +
		"    runs-on: &runner ubuntu-latest\n" +
		"    timeout-minutes: 7\n" +
		"    steps:\n" +
		"      - run: echo base\n" +
		"  build:\n" +
		"    <<: &shared\n" +
		"      timeout-minutes: 7\n" +
		"    runs-on: *runner\n" +
		"    steps:\n" +
		"      - run: echo build\n"

	got := unparse(t, yaml)

	if strings.Contains(got, "*runner") || strings.Contains(got, "<<") {
		t.Errorf("an alias or a merge key reached the output unresolved:\n%s", got)
	}

	if count := strings.Count(got, "timeout_minutes = 7"); count != 2 {
		t.Errorf("want the merged timeout on both jobs, got %d:\n%s", count, got)
	}

	if count := strings.Count(got, `runners = "ubuntu-latest"`); count != 2 {
		t.Errorf("want the aliased runner on both jobs, got %d:\n%s", count, got)
	}
}

// A null decodes to no value at all, and the writer has to say so rather than
// invent an empty string.
func TestNullStaysNull(t *testing.T) {
	for _, value := range []string{"~", "null", ""} {
		if got := unparseEnvValue(t, value); !strings.Contains(got, "value = null") {
			t.Errorf("%q did not come back as null:\n%s", value, got)
		}
	}
}

// unparseEnvValue runs a workflow carrying value as a step environment value
// and returns the HCL.
func unparseEnvValue(t *testing.T, value string) string {
	t.Helper()

	return unparse(t, "on: push\n"+
		"jobs:\n"+
		"  build:\n"+
		"    runs-on: ubuntu-latest\n"+
		"    steps:\n"+
		"      - run: echo hi\n"+
		"        env:\n"+
		"          V: "+value+"\n")
}

// unparse writes the YAML to a file, unparses it and returns the HCL.
func unparse(t *testing.T, yaml string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	output := filepath.Join(dir, "out")

	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: path, OutputDirectory: output}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}

	var hcl strings.Builder

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(output, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}

		hcl.Write(content)
	}

	return hcl.String()
}

// unparseErr writes the YAML to a file and returns the error unparsing it
// gives, if any.
func unparseErr(t *testing.T, yaml string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")

	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	return New().Unparse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})
}

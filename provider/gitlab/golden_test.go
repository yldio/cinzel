// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

func TestParseGoldenFixtures(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "fixtures", "pipelines", "*.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".hcl")
		t.Run(name, func(t *testing.T) {
			outDir := t.TempDir()

			if err := New().Parse(provider.ProviderOps{File: input, OutputDirectory: outDir}); err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(filepath.Join(outDir, ".gitlab-ci.yml"))
			if err != nil {
				t.Fatal(err)
			}

			expected, err := os.ReadFile(strings.TrimSuffix(input, ".hcl") + ".golden.yaml")
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLSemanticEqual(t, got, expected)
		})
	}
}

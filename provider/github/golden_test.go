// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/yldio/cinzel/provider"
)

func TestParseGoldenFixtures(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		expected   string
	}{
		{
			name:       "basic workflow",
			inputFile:  "testdata/fixtures/workflows/basic_workflow.hcl",
			outputFile: "basic_workflow.yaml",
			expected:   "testdata/fixtures/workflows/basic_workflow.golden.yaml",
		},
		{
			name:       "workflow call",
			inputFile:  "testdata/fixtures/workflows/workflow_call.hcl",
			outputFile: "workflow_call.yaml",
			expected:   "testdata/fixtures/workflows/workflow_call.golden.yaml",
		},
		{
			name:       "reusable job",
			inputFile:  "testdata/fixtures/workflows/reusable_job.hcl",
			outputFile: "reusable_job.yaml",
			expected:   "testdata/fixtures/workflows/reusable_job.golden.yaml",
		},
		{
			name:       "permissions concurrency defaults environment",
			inputFile:  "testdata/fixtures/workflows/permissions_concurrency.hcl",
			outputFile: "permissions_concurrency.yaml",
			expected:   "testdata/fixtures/workflows/permissions_concurrency.golden.yaml",
		},
		{
			name:       "container and services",
			inputFile:  "testdata/fixtures/workflows/container_services.hcl",
			outputFile: "container_services.yaml",
			expected:   "testdata/fixtures/workflows/container_services.golden.yaml",
		},
		{
			name:       "step only",
			inputFile:  "testdata/fixtures/workflows/step_only.hcl",
			outputFile: "step_only.yaml",
			expected:   "testdata/fixtures/workflows/step_only.golden.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := t.TempDir()

			p := New()
			err := p.Parse(provider.ProviderOps{
				File:            filepath.Join(".", tt.inputFile),
				OutputDirectory: outDir,
			})
			if err != nil {
				t.Fatal(err)
			}

			gotBytes, err := os.ReadFile(filepath.Join(outDir, tt.outputFile))
			if err != nil {
				t.Fatal(err)
			}

			expBytes, err := os.ReadFile(filepath.Join(".", tt.expected))
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLSemanticEqual(t, gotBytes, expBytes)
		})
	}
}

func TestParseActionGoldenFixtures(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		expected   string
	}{
		{
			name:       "composite action",
			inputFile:  "testdata/fixtures/actions/composite_action.hcl",
			outputFile: "composite_action/action.yml",
			expected:   "testdata/fixtures/actions/composite_action.golden.yaml",
		},
		{
			name:       "node action",
			inputFile:  "testdata/fixtures/actions/node_action.hcl",
			outputFile: "node_action/action.yml",
			expected:   "testdata/fixtures/actions/node_action.golden.yaml",
		},
		{
			name:       "docker action",
			inputFile:  "testdata/fixtures/actions/docker_action.hcl",
			outputFile: "docker_action/action.yml",
			expected:   "testdata/fixtures/actions/docker_action.golden.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := t.TempDir()

			p := New()
			err := p.Parse(provider.ProviderOps{
				File:            filepath.Join(".", tt.inputFile),
				OutputDirectory: outDir,
			})
			if err != nil {
				t.Fatal(err)
			}

			gotBytes, err := os.ReadFile(filepath.Join(outDir, tt.outputFile))
			if err != nil {
				t.Fatal(err)
			}

			expBytes, err := os.ReadFile(filepath.Join(".", tt.expected))
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLSemanticEqual(t, gotBytes, expBytes)
		})
	}
}

func assertYAMLSemanticEqual(t *testing.T, got []byte, expected []byte) {
	t.Helper()

	var gotValue any

	if err := yaml.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("failed to decode generated yaml: %v", err)
	}

	var expectedValue any

	if err := yaml.Unmarshal(expected, &expectedValue); err != nil {
		t.Fatalf("failed to decode expected yaml: %v", err)
	}

	if !reflect.DeepEqual(gotValue, expectedValue) {
		t.Fatalf("generated yaml does not match expected\n--- got ---\n%s\n--- expected ---\n%s", string(got), string(expected))
	}
}

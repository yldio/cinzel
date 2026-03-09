// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/cinzel/provider"
)

func mustParseYAMLDoc(b *testing.B, content []byte) map[string]any {
	b.Helper()
	doc, err := parseYAMLDocument(content)
	if err != nil {
		b.Fatal(err)
	}

	return doc
}

func BenchmarkParseWorkflow(b *testing.B) {
	p := New()
	input := filepath.Join("testdata", "fixtures", "workflows", "basic_workflow.hcl")
	outputDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.Parse(provider.ProviderOps{File: input, OutputDirectory: outputDir}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnparseWorkflow(b *testing.B) {
	p := New()
	input := filepath.Join("testdata", "fixtures", "workflows", "basic_workflow.golden.yaml")
	outputDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.Unparse(provider.ProviderOps{File: input, OutputDirectory: outputDir}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundtripWorkflow(b *testing.B) {
	p := New()
	inputHCL := filepath.Join("testdata", "fixtures", "workflows", "workflow_call.hcl")
	parseDir := b.TempDir()
	unparseDir := b.TempDir()
	parseAgainDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.Parse(provider.ProviderOps{File: inputHCL, OutputDirectory: parseDir}); err != nil {
			b.Fatal(err)
		}

		yamlFile := filepath.Join(parseDir, "workflow_call.yaml")

		if err := p.Unparse(provider.ProviderOps{File: yamlFile, OutputDirectory: unparseDir}); err != nil {
			b.Fatal(err)
		}

		hclFile := filepath.Join(unparseDir, "workflow_call.hcl")

		if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parseAgainDir}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseWorkflowInMemory(b *testing.B) {
	input := filepath.Join("testdata", "fixtures", "workflows", "workflow_call.hcl")
	content, err := os.ReadFile(input)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		parser := hclparse.NewParser()
		file, diags := parser.ParseHCL(content, input)

		if diags.HasErrors() {
			b.Fatal(diags.Error())
		}

		if _, _, _, err := parseHCLToWorkflows(file.Body); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnparseWorkflowInMemory(b *testing.B) {
	input := filepath.Join("testdata", "fixtures", "workflows", "workflow_call.golden.yaml")
	content, err := os.ReadFile(input)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		doc, err := classifyWorkflowDocument(mustParseYAMLDoc(b, content))
		if err != nil {
			b.Fatal(err)
		}

		if doc == nil {
			b.Fatal("expected workflow document")
		}

		if _, err := workflowToHCL(*doc, "workflow_call"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWorkflowToHCLInMemory(b *testing.B) {
	input := filepath.Join("testdata", "fixtures", "workflows", "workflow_call.golden.yaml")
	content, err := os.ReadFile(input)
	if err != nil {
		b.Fatal(err)
	}

	doc, err := classifyWorkflowDocument(mustParseYAMLDoc(b, content))
	if err != nil {
		b.Fatal(err)
	}

	if doc == nil {
		b.Fatal("expected workflow document")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := workflowToHCL(*doc, "workflow_call"); err != nil {
			b.Fatal(err)
		}
	}
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

func BenchmarkParsePipeline(b *testing.B) {
	p := New()
	input := filepath.Join("testdata", "fixtures", "pipelines", "basic_pipeline.hcl")
	outputDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.Parse(provider.ProviderOps{File: input, OutputDirectory: outputDir}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnparsePipeline(b *testing.B) {
	p := New()
	input := filepath.Join("testdata", "fixtures", "pipelines", "basic_pipeline.golden.yaml")
	outputDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.Unparse(provider.ProviderOps{File: input, OutputDirectory: outputDir}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundtripPipeline(b *testing.B) {
	p := New()
	inputHCL := filepath.Join("testdata", "fixtures", "pipelines", "depends_on.hcl")
	parseDir := b.TempDir()
	unparseDir := b.TempDir()
	parseAgainDir := b.TempDir()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := p.Parse(provider.ProviderOps{File: inputHCL, OutputDirectory: parseDir}); err != nil {
			b.Fatal(err)
		}

		yamlFile := filepath.Join(parseDir, ".gitlab-ci.yml")

		if err := p.Unparse(provider.ProviderOps{File: yamlFile, OutputDirectory: unparseDir}); err != nil {
			b.Fatal(err)
		}

		hclFile := filepath.Join(unparseDir, ".gitlab-ci.hcl")

		if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parseAgainDir}); err != nil {
			b.Fatal(err)
		}
	}
}

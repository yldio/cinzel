// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yldio/cinzel/internal/fsutil"
	"github.com/yldio/cinzel/provider"
)

const (
	defaultParseOutputDirectory   = "."
	defaultUnparseOutputDirectory = "./cinzel"

	providerName = "gitlab"
	providerDesc = "GitLab CI/CD Pipelines https://about.gitlab.com/stages-devops-lifecycle/continuous-integration/"
	parseDesc    = "Convert HCL definitions to GitLab CI/CD YAML"
	unparseDesc  = "Convert GitLab CI/CD YAML to HCL definitions"
)

// GitLab is the provider implementation for GitLab CI/CD pipelines.
type GitLab struct{}

// New returns a new GitLab provider instance.
func New() *GitLab { return &GitLab{} }

// GetProviderName returns the provider identifier.
func (p *GitLab) GetProviderName() string { return providerName }

// GetDescription returns a human-readable provider description.
func (p *GitLab) GetDescription() string { return providerDesc }

// GetParseDescription returns a parse operation description.
func (p *GitLab) GetParseDescription() string { return parseDesc }

// GetUnparseDescription returns an unparse operation description.
func (p *GitLab) GetUnparseDescription() string { return unparseDesc }

// Parse converts HCL pipeline definitions to GitLab YAML.
func (p *GitLab) Parse(opts provider.ProviderOps) error {
	inputPath, err := resolveInputPath(opts)
	if err != nil {
		return err
	}

	body, err := fsutil.ParseHCLInput(inputPath, opts.Recursive)
	if err != nil {
		return err
	}

	pipeline, err := parseHCLToPipeline(body)
	if err != nil {
		return err
	}

	outputBytes, err := marshalPipelineYAML(pipeline)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(resolveParseOutputDirectory(opts), ".gitlab-ci.yml")
	if opts.DryRun {
		fmt.Printf("# file: %s\n", outputPath)
		fmt.Println(string(outputBytes))
		return nil
	}

	if err := fsutil.WriteFile(outputPath, outputBytes); err != nil {
		return err
	}

	return nil
}

// Unparse converts GitLab YAML to HCL pipeline definitions.
func (p *GitLab) Unparse(opts provider.ProviderOps) error {
	inputPath, err := resolveInputPath(opts)
	if err != nil {
		return err
	}

	files, err := fsutil.ListFilesWithExtensions(inputPath, opts.Recursive, ".yaml", ".yml")
	if err != nil {
		return err
	}

	outputDir := resolveUnparseOutputDirectory(opts)
	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		doc, err := parseYAMLDocument(yamlBytes)
		if err != nil {
			return err
		}

		if !classifyPipelineDocument(doc) {
			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		hclBytes, err := pipelineToHCL(doc, baseName)
		if err != nil {
			return fmt.Errorf("error in file '%s': %w", file, err)
		}

		outputPath := filepath.Join(outputDir, baseName+".hcl")
		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(hclBytes))
			continue
		}

		if err := fsutil.WriteFile(outputPath, hclBytes); err != nil {
			return err
		}
	}

	return nil
}

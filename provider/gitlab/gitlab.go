// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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

	body, sources, err := fsutil.ParseHCLInput(inputPath, opts.Recursive)
	if err != nil {
		return err
	}

	pipeline, comments, err := parseHCLToPipeline(body, sources)
	if err != nil {
		return err
	}

	// Nothing was declared. Writing here would leave a "{}" document behind.
	if len(pipeline) == 0 {
		return errNoDefinitions
	}

	outputBytes, err := marshalPipelineYAML(pipeline, comments)
	if err != nil {
		return err
	}

	outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName)

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

	// Nothing was read. Reporting success here says the input was converted
	// when it was not: pointing at the wrong directory, or forgetting
	// --recursive with the pipeline a level down, both land exactly here.
	if len(files) == 0 {
		return errNoYAMLFiles
	}

	outputDir := resolveUnparseOutputDirectory(opts)

	// Files were read but none of them held a pipeline, so again nothing was
	// written. The same silence hides the same mistake. A dry run counts: it
	// found the pipeline and only skipped the write it was told to skip.
	found := false

	// The output name comes from the input's basename, so a recursive run over
	// two directories each holding a ".gitlab-ci.yml" aimed both at one path.
	takenNames := make(map[string]struct{}, len(files))

	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// Named, the way the conversion error below is. A directory run reads
		// every ".yaml" in the tree, and the parser's complaint carries a line
		// and column but no file: on its own it says a pipeline somewhere
		// failed to parse without saying which.
		doc, err := parseYAMLDocument(yamlBytes)
		if err != nil {
			return fmt.Errorf("error in file '%s': %w", file, err)
		}

		if !classifyPipelineDocument(doc) {
			continue
		}

		found = true

		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		hclBytes, err := pipelineToHCL(doc, baseName, documentComments(yamlBytes))
		if err != nil {
			return fmt.Errorf("error in file '%s': %w", file, err)
		}

		outputPath := filepath.Join(outputDir, fsutil.UniqueOutputName(takenNames, baseName)+".hcl")

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(hclBytes))
			continue
		}

		if err := fsutil.WriteFile(outputPath, hclBytes); err != nil {
			return err
		}
	}

	if !found {
		return errNoDefinitions
	}

	return nil
}

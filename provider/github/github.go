// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/fsutil"
	"github.com/yldio/cinzel/provider"
)

const (
	defaultParseOutputDirectory   = ".github/workflows"
	defaultUnparseOutputDirectory = "./cinzel"

	providerName = "github"
	providerDesc = "GitHub Actions https://github.com/features/actions"
	parseDesc    = "Convert HCL definitions to GitHub Actions YAML"
	unparseDesc  = "Convert GitHub Actions YAML to HCL definitions"
)

// GitHub is the provider implementation for GitHub Actions workflows.
type GitHub struct{}

// New returns a new GitHub provider instance.
func New() *GitHub { return &GitHub{} }

// GetProviderName returns the provider identifier.
func (p *GitHub) GetProviderName() string { return providerName }

// GetDescription returns a human-readable description of the provider.
func (p *GitHub) GetDescription() string { return providerDesc }

// GetParseDescription returns a description of the parse (HCL to YAML) operation.
func (p *GitHub) GetParseDescription() string { return parseDesc }

// GetUnparseDescription returns a description of the unparse (YAML to HCL) operation.
func (p *GitHub) GetUnparseDescription() string { return unparseDesc }

// Parse converts HCL workflow definitions into GitHub Actions YAML files.
func (p *GitHub) Parse(opts provider.ProviderOps) error {
	inputPath, err := resolveInputPath(opts)
	if err != nil {

		return err
	}

	body, err := fsutil.ParseHCLInput(inputPath, opts.Recursive)
	if err != nil {

		return err
	}

	workflows, stepMap, actions, err := parseHCLToWorkflows(body)
	if err != nil {

		return err
	}

	outputDir := resolveParseOutputDirectory(opts)

	if len(workflows) == 0 && len(actions) == 0 {
		outputBytes, err := yaml.Marshal(stepMap)
		if err != nil {

			return err
		}

		outputPath := filepath.Join(outputDir, resolveParseFilename(opts))

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(outputBytes))

			return nil
		}

		return fsutil.WriteFile(outputPath, outputBytes)
	}

	for _, workflowFile := range workflows {
		outputBytes, err := marshalWorkflowYAML(workflowFile.Content)
		if err != nil {

			return err
		}

		outputPath := filepath.Join(outputDir, workflowFile.Filename+".yaml")

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(outputBytes))
			continue
		}

		if err := fsutil.WriteFile(outputPath, outputBytes); err != nil {

			return err
		}
	}

	for _, actionFile := range actions {
		outputBytes, err := marshalWorkflowYAML(actionFile.Content)
		if err != nil {

			return err
		}

		outputPath := filepath.Join(outputDir, actionFile.Filename, "action.yml")

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(outputBytes))
			continue
		}

		if err := fsutil.WriteFile(outputPath, outputBytes); err != nil {

			return err
		}
	}

	return nil
}

// Unparse converts GitHub Actions YAML files into HCL definitions.
func (p *GitHub) Unparse(opts provider.ProviderOps) error {
	inputPath, err := resolveInputPath(opts)
	if err != nil {

		return err
	}

	files, err := fsutil.ListFilesWithExtensions(inputPath, opts.Recursive, ".yaml", ".yml")
	if err != nil {

		return err
	}

	if len(files) == 0 {

		return errNoYAMLFiles
	}

	outputDir := resolveUnparseOutputDirectory(opts)

	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {

			return err
		}

		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		hclBytes, err := unparseYAMLFile(yamlBytes, baseName)
		if err != nil {

			return fmt.Errorf("error in file '%s': %w", file, err)
		}

		if hclBytes == nil {
			continue
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

// unparseYAMLFile converts a YAML file to HCL bytes, detecting whether
// the document is a workflow, action, or step-only file. Returns nil if
// the document is empty or unrecognized.
func unparseYAMLFile(yamlBytes []byte, baseName string) ([]byte, error) {
	doc, err := parseYAMLDocument(yamlBytes)
	if err != nil {

		return nil, err
	}

	workflowDoc, err := classifyWorkflowDocument(doc)
	if err != nil {

		return nil, err
	}

	if workflowDoc != nil {

		return workflowToHCL(*workflowDoc, baseName)
	}

	if actionDoc := classifyActionDocument(doc); actionDoc != nil {

		return actionToHCL(actionDoc, baseName)
	}

	steps, err := parseStepsFromYAML(yamlBytes)
	if err != nil {

		return nil, err
	}

	if len(steps) == 0 {

		return nil, nil
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()

	for _, s := range steps {

		if err := s.Decode(body, "step"); err != nil {

			return nil, err
		}
	}

	return hclwrite.Format(f.Bytes()), nil
}

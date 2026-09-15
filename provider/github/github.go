// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		// Nothing was declared. Writing here would leave a "{}" document
		// behind and, worse, make it the only current output, so the prune
		// below would delete every other generated file in the directory.
		if len(stepMap) == 0 {
			return errNoDefinitions
		}

		outputBytes, err := marshalStepsYAML(stepMap)
		if err != nil {
			return err
		}

		outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName)

		outputPath := filepath.Join(outputDir, resolveParseFilename(opts))

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(outputBytes))

			return nil
		}

		if err := fsutil.WriteFile(outputPath, outputBytes); err != nil {
			return err
		}

		currentWorkflowOutputs := map[string]struct{}{}
		currentWorkflowOutputs[filepath.Clean(outputPath)] = struct{}{}

		return fsutil.PruneStaleGeneratedYAML(outputDir, currentWorkflowOutputs, providerName)
	}

	// Every file this run writes has to be recorded, actions included: the
	// prune below walks the whole tree, and anything it does not find here is
	// read as stale and removed.
	currentOutputs := make(map[string]struct{}, len(workflows)+len(actions))

	for _, workflowFile := range workflows {
		outputBytes, err := marshalWorkflowYAML(workflowFile.Content, workflowFile.JobOrder)
		if err != nil {
			return err
		}

		outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName)

		outputPath := filepath.Join(outputDir, workflowFile.Filename+workflowExt(opts))
		currentOutputs[filepath.Clean(outputPath)] = struct{}{}

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
		outputBytes, err := marshalWorkflowYAML(actionFile.Content, nil)
		if err != nil {
			return err
		}

		outputPath := filepath.Join(outputDir, actionFile.Filename, "action.yml")
		currentOutputs[filepath.Clean(outputPath)] = struct{}{}

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(outputBytes))
			continue
		}

		if err := fsutil.WriteFile(outputPath, outputBytes); err != nil {
			return err
		}
	}

	if opts.DryRun {
		return nil
	}

	return fsutil.PruneStaleGeneratedYAML(outputDir, currentOutputs, providerName)
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

		hclBytes, name, err := unparseYAMLFile(yamlBytes, baseName, actionNameFor(file))
		if err != nil {
			return fmt.Errorf("error in file '%s': %w", file, err)
		}

		if hclBytes == nil {
			continue
		}

		outputPath := filepath.Join(outputDir, name+".hcl")

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
func unparseYAMLFile(yamlBytes []byte, baseName, actionName string) ([]byte, string, error) {
	doc, jobOrder, err := parseYAMLDocument(yamlBytes)
	if err != nil {
		return nil, "", err
	}

	// A file holding no document at all is skipped rather than rejected, the
	// same as one holding an empty document. Otherwise a single stray empty
	// file aborts a whole directory run before the real files are reached.
	if doc == nil {
		return nil, "", nil
	}

	workflowDoc, err := classifyWorkflowDocument(doc)
	if err != nil {
		return nil, "", err
	}

	if workflowDoc != nil {
		out, err := workflowToHCL(*workflowDoc, baseName, jobOrder)

		return out, baseName, err
	}

	if actionDoc := classifyActionDocument(doc); actionDoc != nil {
		out, err := actionToHCL(actionDoc, actionName)

		return out, actionName, err
	}

	steps, err := parseStepsFromYAML(yamlBytes)
	if err != nil {
		return nil, "", err
	}

	if len(steps) == 0 {
		return nil, "", nil
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()

	for _, s := range steps {
		if err := s.Decode(body, "step"); err != nil {
			return nil, "", err
		}
	}

	return unescapeHCLUnicode(hclwrite.Format(f.Bytes())), baseName, nil
}

// actionNameFor picks the name an action is known by. An action lives at
// <name>/action.yml, so the directory holding the file is its identity and
// the basename is the same word for every action there is. Naming the output
// after the basename dropped that identity: the action came back called
// "action", and a second one in the same run wrote over the first.
//
// Only the fixed filenames GitHub reads are treated this way. A document that
// happens to be an action but sits under some other name is left with that
// name, which is the only thing left to go on.
func actionNameFor(file string) string {
	base := strings.ToLower(filepath.Base(file))

	if base != "action.yml" && base != "action.yaml" {
		return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}

	dir := filepath.Base(filepath.Dir(file))

	// A file directly under "." or the filesystem root has no directory name
	// to take. Nothing is better than the basename there.
	if dir == "." || dir == string(os.PathSeparator) || dir == "" {
		return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}

	return dir
}

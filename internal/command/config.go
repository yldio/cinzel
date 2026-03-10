// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/provider"
	"gopkg.in/yaml.v3"
)

const configFilename = ".cinzelrc.yaml"

type providerCommandConfig struct {
	file            string
	hasFile         bool
	directory       string
	hasDirectory    bool
	outputDirectory string
	hasOutputDir    bool
}

func toProviderOpts(cmd *cli.Command, providerName string, commandName string) (provider.ProviderOps, []string, error) {
	opts := provider.ProviderOps{
		File:            cmd.String("file"),
		Directory:       cmd.String("directory"),
		OutputDirectory: cmd.String("output-directory"),
		Recursive:       cmd.Bool("recursive"),
		DryRun:          cmd.Bool("dry-run"),
	}

	conf, warnings, err := loadProviderCommandConfig(configFilename, providerName, commandName)
	if err != nil {
		return provider.ProviderOps{}, nil, err
	}

	if !cmd.IsSet("output-directory") && conf.hasOutputDir {
		opts.OutputDirectory = conf.outputDirectory
	}

	hasCLIFileInput := cmd.IsSet("file") || cmd.IsSet("directory")

	if !hasCLIFileInput {
		if conf.hasFile {
			opts.File = conf.file
		}

		if conf.hasDirectory {
			opts.Directory = conf.directory
		}
	}

	return opts, warnings, nil
}

func loadProviderCommandConfig(path string, providerName string, commandName string) (providerCommandConfig, []string, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return providerCommandConfig{}, nil, nil
		}

		return providerCommandConfig{}, nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var doc yaml.Node

	if err := yaml.Unmarshal(configBytes, &doc); err != nil {
		return providerCommandConfig{}, nil, fmt.Errorf("invalid %s: %w", path, err)
	}

	if len(doc.Content) == 0 {
		return providerCommandConfig{}, nil, nil
	}

	root := doc.Content[0]

	if root.Kind != yaml.MappingNode {
		return providerCommandConfig{}, nil, fmt.Errorf("%s must contain a YAML mapping", path)
	}

	warnings := make([]string, 0)
	providerNode := findMappingValue(root, providerName)

	if providerNode == nil {
		return providerCommandConfig{}, warnings, nil
	}

	if providerNode.Kind != yaml.MappingNode {
		return providerCommandConfig{}, nil, fmt.Errorf("%s.%s must be a mapping", path, providerName)
	}

	for i := 0; i < len(providerNode.Content); i += 2 {
		k := providerNode.Content[i].Value

		if k != "parse" && k != "unparse" {
			warnings = append(warnings, fmt.Sprintf("%s.%s.%s: unknown key", path, providerName, k))
		}
	}

	commandNode := findMappingValue(providerNode, commandName)

	if commandNode == nil {
		sort.Strings(warnings)

		return providerCommandConfig{}, warnings, nil
	}

	if commandNode.Kind != yaml.MappingNode {
		return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s must be a mapping", path, providerName, commandName)
	}

	config := providerCommandConfig{}

	for i := 0; i < len(commandNode.Content); i += 2 {
		keyNode := commandNode.Content[i]
		valueNode := commandNode.Content[i+1]

		switch keyNode.Value {
		case "file":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!str" {
				return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s.file must be string", path, providerName, commandName)
			}
			config.file = valueNode.Value
			config.hasFile = true
		case "directory":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!str" {
				return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s.directory must be string", path, providerName, commandName)
			}
			config.directory = valueNode.Value
			config.hasDirectory = true
		case "output-directory":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!str" {
				return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s.output-directory must be string", path, providerName, commandName)
			}
			config.outputDirectory = valueNode.Value
			config.hasOutputDir = true
		case "single-file":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!bool" {
				return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s.single-file must be boolean", path, providerName, commandName)
			}
		case "filename":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!str" {
				return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s.filename must be string", path, providerName, commandName)
			}
		default:
			warnings = append(warnings, fmt.Sprintf("%s.%s.%s.%s: unknown key", path, providerName, commandName, keyNode.Value))
		}
	}

	if config.hasFile && config.hasDirectory {
		return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s cannot set both file and directory", path, providerName, commandName)
	}

	sort.Strings(warnings)

	return config, warnings, nil
}

func findMappingValue(n *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}

	return nil
}

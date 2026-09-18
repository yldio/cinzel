// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	yml             bool
	hasYML          bool
}

func toProviderOpts(cmd *cli.Command, providerName string, commandName string) (provider.ProviderOps, []string, error) {
	opts := provider.ProviderOps{
		File:            cmd.String("file"),
		Directory:       cmd.String("directory"),
		OutputDirectory: cmd.String("output-directory"),
		Recursive:       cmd.Bool("recursive"),
		DryRun:          cmd.Bool("dry-run"),
		YML:             cmd.Bool("yml"),
	}

	conf, warnings, err := loadProviderCommandConfig(configFilename, providerName, commandName)
	if err != nil {
		return provider.ProviderOps{}, nil, err
	}

	if !cmd.IsSet("output-directory") && conf.hasOutputDir {
		opts.OutputDirectory = conf.outputDirectory
	}

	if !cmd.IsSet("yml") && conf.hasYML {
		opts.YML = conf.yml
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
			value, err := pathValue(valueNode, path, providerName, commandName, keyNode.Value)
			if err != nil {
				return providerCommandConfig{}, nil, err
			}
			config.file = value
			config.hasFile = true
		case "directory":
			value, err := pathValue(valueNode, path, providerName, commandName, keyNode.Value)
			if err != nil {
				return providerCommandConfig{}, nil, err
			}
			config.directory = value
			config.hasDirectory = true
		case "output-directory":
			value, err := pathValue(valueNode, path, providerName, commandName, keyNode.Value)
			if err != nil {
				return providerCommandConfig{}, nil, err
			}
			config.outputDirectory = value
			config.hasOutputDir = true
		case "yml":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!bool" {
				return providerCommandConfig{}, nil, fmt.Errorf("%s.%s.%s.yml must be a boolean", path, providerName, commandName)
			}
			config.yml = valueNode.Value == "true"
			config.hasYML = true
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

// pathValue reads one path out of the configuration, refusing anything that
// will not travel.
//
// The file is committed to git and read on every machine that checks the repo
// out, so a path in it has to mean the same thing on all of them. An absolute
// path names one machine's disk, and a leading "~" names one machine's user.
// Both are rejected here rather than failing later as a missing file, which
// says nothing about why.
//
// Forward slashes are turned into whatever the running OS separates with, so
// one committed spelling works everywhere. The reverse is not done: a config
// written on Windows with backslashes is a path a POSIX reader takes as one
// filename with backslashes in it, and quietly rewriting it would guess at
// which the author meant.
func pathValue(node *yaml.Node, path, providerName, commandName, key string) (string, error) {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return "", fmt.Errorf("%s.%s.%s.%s must be string", path, providerName, commandName, key)
	}

	if filepath.IsAbs(node.Value) || strings.HasPrefix(node.Value, "~") {
		return "", fmt.Errorf(
			"%s.%s.%s.%s must be a relative path (got %q): the file is shared, so a path in it has to work on every machine",
			path, providerName, commandName, key, node.Value,
		)
	}

	return filepath.FromSlash(node.Value), nil
}

func findMappingValue(n *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}

	return nil
}

// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/pin"
	"github.com/yldio/cinzel/provider"
)

func (cmd *Cli) upgradeCommand() *cli.Command {
	return &cli.Command{
		Name:  "upgrade",
		Usage: "Upgrade GitHub Actions to their latest versions and pin to SHAs",
		Action: func(ctx context.Context, c *cli.Command) error {
			filePath := c.String("file")
			dirPath := c.String("directory")
			dryRun := c.Bool("dry-run")
			parse := c.Bool("parse")

			if filePath == "" && dirPath == "" {
				dirPath = "cinzel"
			}

			if filePath != "" {
				if err := validateRelativePath(filePath); err != nil {
					return fmt.Errorf("--file: %w", err)
				}
			}

			if dirPath != "" {
				if err := validateRelativePath(dirPath); err != nil {
					return fmt.Errorf("--directory: %w", err)
				}
			}

			// No cache wrapper — upgrade checks latest releases which should not
			// be served from a 24h cache.
			resolver := pin.NewGitHubResolver("")

			var results []pin.UpgradeResult

			var err error

			if filePath != "" {
				results, err = pin.UpgradeFile(ctx, filePath, resolver, cmd.Writer, dryRun)
			} else {
				results, err = pin.UpgradeDirectory(ctx, dirPath, resolver, cmd.Writer, dryRun)
			}

			if err != nil {
				return err
			}

			cmd.printUpgradeSummary(results)

			if !parse || dryRun {
				return nil
			}

			// Check if anything was actually upgraded.
			upgraded := false

			for _, r := range results {
				if r.Error == nil && !r.WasCurrent {
					upgraded = true

					break
				}
			}

			if !upgraded {
				return nil
			}

			parseDir := dirPath
			if filePath != "" {
				// If a single file was upgraded, parse its parent directory.
				parseDir = filepath.Dir(filePath)
			}

			outputDir := c.String("output-directory")
			if outputDir != "" {
				if err := validateRelativePath(outputDir); err != nil {
					return fmt.Errorf("--output-directory: %w", err)
				}
			}

			return cmd.runParseGitHub(parseDir, outputDir)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "Upgrade actions in a single HCL `FILE`",
			},
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "Upgrade actions in all HCL files in `DIRECTORY` (default: cinzel)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Value: false,
				Usage: "Show what would be upgraded without writing files",
			},
			&cli.BoolFlag{
				Name:  "parse",
				Value: false,
				Usage: "Regenerate GitHub Actions YAML files after upgrading",
			},
			&cli.StringFlag{
				Name:  "output-directory",
				Value: "",
				Usage: "Output directory for parsed YAML (default: .github/workflows)",
			},
		},
	}
}

func (cmd *Cli) runParseGitHub(inputDir, outputDir string) error {
	p, err := cmd.resolveProvider("github")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.Writer, "\nRegenerating YAML...\n")

	return p.Parse(provider.ProviderOps{
		Directory:       inputDir,
		OutputDirectory: outputDir,
		DryRun:          false,
	})
}

func (cmd *Cli) printUpgradeSummary(results []pin.UpgradeResult) {
	upgraded := 0
	current := 0
	failed := 0

	for _, r := range results {
		switch {
		case r.WasCurrent:
			current++
		case r.Error != nil:
			failed++
		default:
			upgraded++
		}
	}

	_, _ = fmt.Fprintf(cmd.Writer, "\nUpgrade summary: %d upgraded, %d already current, %d failed\n", upgraded, current, failed)
}

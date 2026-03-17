// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/pin"
)

func (cmd *Cli) upgradeCommand() *cli.Command {
	return &cli.Command{
		Name:  "upgrade",
		Usage: "Upgrade GitHub Actions to their latest versions and pin to SHAs",
		Action: func(ctx context.Context, c *cli.Command) error {
			filePath := c.String("file")
			dirPath := c.String("directory")
			dryRun := c.Bool("dry-run")

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

			resolver := pin.NewGitHubResolver("")

			if filePath != "" {
				results, err := pin.UpgradeFile(ctx, filePath, resolver, cmd.Writer, dryRun)
				if err != nil {
					return err
				}

				cmd.printUpgradeSummary(results)

				return nil
			}

			results, err := pin.UpgradeDirectory(ctx, dirPath, resolver, cmd.Writer, dryRun)
			if err != nil {
				return err
			}

			cmd.printUpgradeSummary(results)

			return nil
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
		},
	}
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

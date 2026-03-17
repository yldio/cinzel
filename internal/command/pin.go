// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/pin"
)

func (cmd *Cli) pinCommand() *cli.Command {
	return &cli.Command{
		Name:  "pin",
		Usage: "Resolve GitHub Actions version tags to commit SHAs (no token required for public actions)",
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

			resolver := pin.NewCachedResolver(pin.NewGitHubResolver(""))

			if filePath != "" {
				results, err := pin.PinFile(ctx, filePath, resolver, cmd.Writer, dryRun)
				if err != nil {
					return err
				}

				cmd.printPinSummary(results)

				return nil
			}

			results, err := pin.PinDirectory(ctx, dirPath, resolver, cmd.Writer, dryRun)
			if err != nil {
				return err
			}

			cmd.printPinSummary(results)

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "Pin actions in a single HCL `FILE`",
			},
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "Pin actions in all HCL files in `DIRECTORY` (default: cinzel)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Value: false,
				Usage: "Show what would be pinned without writing files",
			},
		},
	}
}

func (cmd *Cli) printPinSummary(results []pin.PinResult) {
	pinned := 0
	skipped := 0
	failed := 0

	for _, r := range results {
		switch {
		case r.WasAlready:
			skipped++
		case r.Error != nil:
			failed++
		default:
			pinned++
		}
	}

	_, _ = fmt.Fprintf(cmd.Writer, "\nPin summary: %d pinned, %d already pinned, %d failed\n", pinned, skipped, failed)
}

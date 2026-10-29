// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/fsutil"
)

func (cmd *Cli) fmtCommand() *cli.Command {
	return &cli.Command{
		Name:  "fmt",
		Usage: "Rewrite HCL files in the canonical layout",
		Action: func(ctx context.Context, c *cli.Command) error {
			path, flag := c.String("file"), "file"

			if path == "" {
				path, flag = c.String("directory"), "directory"
			}

			if path == "" {
				path = "cinzel"
			}

			if err := validateRelativePath(path); err != nil {
				return fmt.Errorf("--%s: %w", flag, err)
			}

			if flag == "file" {
				if err := requireFileNotDirectory(path); err != nil {
					return err
				}
			}

			files, err := fsutil.ListFilesWithExtensions(path, c.Bool("recursive"), ".hcl")
			if err != nil {
				return err
			}

			return cmd.formatFiles(files, c.Bool("check"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "Format a single HCL `FILE`",
			},
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "Format the HCL files in `DIRECTORY` (default: cinzel)",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Value:   false,
				Usage:   "Reads the directory recursively",
			},
			&cli.BoolFlag{
				Name:  "check",
				Value: false,
				Usage: "Name the files that are not formatted and exit non-zero, writing nothing",
			},
		},
	}
}

// formatFiles rewrites each file, or under check only names the ones that
// would change.
//
// A file is parsed before it is formatted. hclwrite.Format works on tokens
// and not on a parsed document, so it reformats a file it cannot parse as
// readily as one it can, and writing that back leaves broken HCL laid out
// neatly enough to look deliberate.
func (cmd *Cli) formatFiles(files []string, check bool) error {
	changed := 0

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if _, diags := hclsyntax.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1}); diags.HasErrors() {
			return fmt.Errorf("%s: %w", path, cinzelerror.ProcessHCLDiags(diags))
		}

		formatted := hclwrite.Format(content)

		if fsutil.SameContent(content, formatted) {
			continue
		}

		changed++

		_, _ = fmt.Fprintln(cmd.Writer, path)

		if check {
			continue
		}

		if err := fsutil.WriteFile(path, formatted); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(cmd.Writer, "\n%d checked, %d %s\n", len(files), changed, formattedVerb(check))

	if check && changed > 0 {
		return fmt.Errorf("%w: %d of %d", errUnformatted, changed, len(files))
	}

	return nil
}

func formattedVerb(check bool) string {
	if check {
		return "not formatted"
	}

	return "formatted"
}

// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"net/mail"

	"github.com/urfave/cli/v3"
	"github.com/yldio/acto/provider"
)

const copyright = "(c) 2024-2025 YLD Limited"

type Cli struct {
	Cmd *cli.Command
}

func New(version string) *Cli {
	return &Cli{
		Cmd: &cli.Command{
			Version:                version,
			Name:                   "acto",
			Usage:                  "a tool that converts HCL files to your favourite CICD provider.",
			UseShortOptionHandling: true,
			Copyright:              copyright,
			Authors: []any{
				&mail.Address{Name: "Joao Guimaraes", Address: "jccguimaraes@gmail.com"},
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "file",
					Aliases: []string{"f"},
					Value:   "",
					Usage:   "Load HCL from a `FILE`",
				},
				&cli.StringFlag{
					Name:    "directory",
					Aliases: []string{"d"},
					Value:   "",
					Usage:   "Load HCL files from a `DIRECTORY`",
				},
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r"},
					Value:   false,
					Usage:   "Reads the defined directory recursively",
				},
				&cli.BoolFlag{
					Name:  "dry-run",
					Value: false,
					Usage: "Output to stdout",
				},
				&cli.BoolFlag{
					Name:  "override",
					Value: true,
					Usage: "Overrides existing files without prompting the user",
				},
				&cli.BoolFlag{
					Name:    "watch",
					Aliases: []string{"w"},
					Value:   false,
					Usage:   "Watch mode for continuously regenerate the files",
				},
			},
		},
	}
}

func (cmd *Cli) AddCommand(p provider.Provider) {
	cc := &cli.Command{
		Name:  p.GetName(),
		Usage: p.GetDescription(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output-directory",
				Value: "",
				Usage: "Parsed files from HCL or Parsed (unparse) files to HCL are created in `DIRECTORY`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "parse",
				Usage: "Parse HCL files",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					opts := toProviderOpts(cmd)

					if err := p.Parse(opts); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "unparse",
				Usage: "Parse to HCL files",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					opts := toProviderOpts(cmd)

					if err := p.Unparse(opts); err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	cmd.Cmd.Commands = append(cmd.Cmd.Commands, cc)
}

func toProviderOpts(cmd *cli.Command) provider.ProviderOps {
	return provider.ProviderOps{
		File:            cmd.String("file"),
		Directory:       cmd.String("directory"),
		OutputDirectory: cmd.String("output-directory"),
		Recursive:       cmd.Bool("recursive"),
		DryRun:          cmd.Bool("dry-run"),
		Override:        cmd.Bool("override"),
		Watch:           cmd.Bool("watch"),
	}
}

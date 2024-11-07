// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoflag

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/yldio/acto/internal/actoerrors"
)

const gitHubDir = ".github/workflows"
const copyright = "(c) 2024 YLD Limited"

type ActoCli struct {
	InputDirectory  string
	InputFile       string
	OutputDirectory string
	Recursive       bool
	DryRun          bool
	Override        bool
	Watch           bool
	FromActions     bool
}

func NewFlags() *ActoCli {
	return &ActoCli{}
}

func NewCliApp(flags *ActoCli) *cli.App {
	flags.SetHelpTemplate()

	return &cli.App{
		Name:                   "acto",
		Usage:                  "a tool that converts HCL files to GitHub Actions YAML's workflow files.",
		UseShortOptionHandling: true,
		Flags:                  flags.GetFlags(),
		Action: func(cCtx *cli.Context) error {
			return nil
		},
		Copyright: copyright,
	}
}

func (actoCli *ActoCli) GetInputFileFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "file",
		Aliases:     []string{"f"},
		Value:       "",
		DefaultText: "./acto.hcl",
		Destination: &actoCli.InputFile,
		Action: func(ctx *cli.Context, value string) error {
			if _, err := os.Stat(value); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("file `%s` does not exist, %w", value, actoerrors.ErrOpenIssue)
			}
			return nil
		},
		Usage: "Load HCL from a `FILE`",
	}
}

func (actoCli *ActoCli) GetInputDirectoryFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "directory",
		Aliases:     []string{"d"},
		Value:       "",
		Usage:       "Load HCL files from a `DIRECTORY`",
		DefaultText: "./acto",
		Destination: &actoCli.InputDirectory,
		Action: func(ctx *cli.Context, value string) error {
			if _, err := os.Stat(value); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("directory `%s` does not exist, %w", value, actoerrors.ErrOpenIssue)
			}
			return nil
		},
	}
}

func (actoCli *ActoCli) GetOutputDirectoryFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "output-directory",
		Aliases:     []string{"o"},
		Value:       "",
		DefaultText: "`./github/workflows` for YAML and `./acto` for HCL",
		Usage:       "Create YAML files in `DIRECTORY`",
		Destination: &actoCli.OutputDirectory,
		Action: func(ctx *cli.Context, value string) error {
			return nil
		},
	}
}

func (actoCli *ActoCli) GetRecursiveFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "recursive",
		Value:       false,
		Usage:       "Reads the defined directory recursively",
		Destination: &actoCli.Recursive,
		Action: func(ctx *cli.Context, value bool) error {
			return nil
		},
	}
}

func (actoCli *ActoCli) GetDryRunFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "dry-run",
		Value:       false,
		Usage:       "Converts from HCL to YAML but only outputs to stdout",
		Destination: &actoCli.DryRun,
		Action: func(ctx *cli.Context, value bool) error {
			return nil
		},
	}
}

func (actoCli *ActoCli) GetOverrideFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "override",
		Value:       false,
		Usage:       "Overrides existing files without prompting the user",
		Destination: &actoCli.Override,
		Action: func(ctx *cli.Context, value bool) error {
			return nil
		},
	}
}

func (actoCli *ActoCli) GetWatchFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "watch",
		Aliases:     []string{"w"},
		Value:       false,
		Usage:       "Watch mode for continuously regenerate the files",
		Destination: &actoCli.Watch,
		Action: func(ctx *cli.Context, value bool) error {
			return nil
		},
	}
}

func (actoCli *ActoCli) GetFromActionsFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "from-actions",
		Value:       false,
		Usage:       "Converts from YAML to HCL and outputs to defined file or directory",
		Destination: &actoCli.FromActions,
		Action: func(ctx *cli.Context, value bool) error {
			return nil
		},
	}
}

func (actoCli *ActoCli) GetFlags() []cli.Flag {
	return []cli.Flag{
		actoCli.GetInputFileFlag(),
		actoCli.GetInputDirectoryFlag(),
		actoCli.GetOutputDirectoryFlag(),
		actoCli.GetRecursiveFlag(),
		actoCli.GetDryRunFlag(),
		actoCli.GetOverrideFlag(),
		actoCli.GetWatchFlag(),
		actoCli.GetFromActionsFlag(),
	}
}

func (actoCli *ActoCli) SetHelpTemplate() {
	//		cli.AppHelpTemplate = color.GreenString("NAME:") + `
	//	   {{.Name}} - {{.Usage}}{{ "\n" }}
	//
	// ` + color.GreenString("USAGE:") + `
	//
	//	{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
	//	{{if len .Authors}}
	//
	// ` + color.GreenString("AUTHOR:") + `
	//
	//	{{range .Authors}}{{ . }}{{end}}
	//	{{end}}{{if .Commands}}
	//
	// ` + color.GreenString("COMMANDS:") + `
	// {{range .Commands}}{{if not .HideHelp}}{{ "\t" }}{{join .Names ", "}}{{ "\t" }}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
	// ` + color.GreenString("GLOBAL OPTIONS:") + `
	//
	//	{{range .VisibleFlags}}{{.}}
	//	{{end}}{{end}}{{if .Copyright }}
	//
	// ` + color.GreenString("COPYRIGHT:") + `
	//
	//	{{.Copyright}}
	//	{{end}}{{if .Version}}
	//
	// ` + color.GreenString("VERSION:") + `
	//
	//	{{.Version}}
	//	{{end}}
	//
	// `
}

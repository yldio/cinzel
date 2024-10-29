// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoflag

import (
	"flag"
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type Flags struct {
	Directory string
	File      string
	OutputDir string
	Actions   bool
	Recursive bool
	Version   bool
	Help      bool
}

type ActoCli struct {
	InputDirectory  *string
	InputFile       *string
	OutputDirectory *string
	DryRun          *bool
	Override        *bool
	Watch           *bool
}

func New() *Flags {
	directory := flag.String("d", "", "A `directory` where the acto files exist (required if -f is not set).")
	file := flag.String("f", "", "An HCL `file` that contains acto resources (required if -d is not set).")
	outputDir := flag.String("od", "", "Set the `output` directory where to store the converted files. For YAML it defaults to `.github/workflows`, for HCL it defaults to `./acto` (when using the `-actions` flag).")
	actions := flag.Bool("a", false, "Reads all GitHub `actions` and converts them to HCL.\nSpecifing -f will create only one HCL file, specifing -d will create files such as workflows.hcl, jobs.hcl and so on.\nUse the flag `-output-dir` to define the destination folder (defaults to the folder `./acto`).\nUse the flag `-output-file` to define the output file (defaults to `acto.hcl`)")
	recursive := flag.Bool("r", false, "Reads directory recursiveness (valid only if -d is set).")
	help := flag.Bool("h", false, "Show this `help` output.")
	version := flag.Bool("v", false, "Show the current `version` of acto.")
	// dryRun := flag.Bool("dry-run", false, "Show the current `version` of acto.")
	// watch := flag.Bool("dry-run", false, "Show the current `version` of acto.")

	flag.Parse()

	return &Flags{
		Directory: *directory,
		File:      *file,
		OutputDir: *outputDir,
		Actions:   *actions,
		Recursive: *recursive,
		Version:   *version,
		Help:      *help,
	}
}

func (flags *Flags) SetDirectory(directory string) {
	flags.Directory = directory
}

func (flags *Flags) SetFile(file string) {
	flags.File = file
}

func (flags *Flags) SetRecursive(recursive bool) {
	flags.Recursive = recursive
}

func (flags *Flags) SetActions(actions bool) {
	flags.Actions = actions
}

func (flags *Flags) GetUsage() {
	flag.Usage()
}

func NewFlags() *ActoCli {
	return &ActoCli{}
}

func NewCliApp() *cli.App {
	flags := NewFlags()

	flags.SetHelpTemplate()

	return &cli.App{
		Name:                   "acto",
		Usage:                  "a tool that converts HCL files to GitHub Actions YAML's workflow files.",
		UseShortOptionHandling: true,
		Flags:                  flags.GetFlags(),
	}
}

func (actoCli *ActoCli) GetInputFileFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "file",
		Aliases:     []string{"f"},
		Value:       "",
		DefaultText: "./acto.hcl",
		Action: func(ctx *cli.Context, action string) error {
			fmt.Println("action", action)
			return nil
		},
		Usage:       "Load HCL from a `FILE`",
		Destination: actoCli.InputFile,
	}
}

func (actoCli *ActoCli) GetInputDirectoryFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "directory",
		Aliases:     []string{"d"},
		Value:       "",
		Usage:       "Load HCL from a `DIRECTORY`",
		DefaultText: "./acto",
		Action: func(ctx *cli.Context, action string) error {
			fmt.Println("action", action)
			return nil
		},
		Destination: actoCli.InputDirectory,
	}
}

func (actoCli *ActoCli) GetOutputDirectoryFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "output-directory",
		Aliases:     []string{"o"},
		Value:       "",
		DefaultText: "`./github/workflows` for YAML and `./acto` for HCL",
		Usage:       "Create YAML files in `DIRECTORY`",
		Destination: actoCli.OutputDirectory,
	}
}

func (actoCli *ActoCli) GetDryRunFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "dry-run",
		Value:       false,
		Usage:       "Converts from HCL to YAML but only outputs to stdout",
		Destination: actoCli.DryRun,
	}
}

func (actoCli *ActoCli) GetOverrideFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "override",
		Value:       false,
		Usage:       "Overrides existing files without prompting the user",
		Destination: actoCli.Override,
	}
}

func (actoCli *ActoCli) GeWatchFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "watch",
		Value:       false,
		Usage:       "Watch mode for continuously regenerate the files",
		Destination: actoCli.Watch,
	}
}

func (actoCli *ActoCli) GetFlags() []cli.Flag {
	return []cli.Flag{
		actoCli.GetInputFileFlag(),
		actoCli.GetInputDirectoryFlag(),
		actoCli.GetOutputDirectoryFlag(),
		actoCli.GetDryRunFlag(),
		actoCli.GetOverrideFlag(),
		actoCli.GeWatchFlag(),
	}
}

func (actoCli *ActoCli) SetHelpTemplate() {
	cli.AppHelpTemplate = color.GreenString("NAME:") + `
   {{.Name}} - {{.Usage}}{{ "\n" }}
` + color.GreenString("USAGE:") + `
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
` + color.GreenString("AUTHOR:") + `
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
` + color.GreenString("COMMANDS:") + `
{{range .Commands}}{{if not .HideHelp}}{{ "\t" }}{{join .Names ", "}}{{ "\t" }}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
` + color.GreenString("GLOBAL OPTIONS:") + `
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
` + color.GreenString("COPYRIGHT:") + `
   {{.Copyright}}
   {{end}}{{if .Version}}
` + color.GreenString("VERSION:") + `
   {{.Version}}
   {{end}}
`
}

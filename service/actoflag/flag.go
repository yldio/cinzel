// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoflag

import "flag"

type Flags struct {
	Directory string
	File      string
	Recursive bool
	Version   bool
	Help      bool
}

func New() *Flags {
	directory := flag.String("dir", "", "A `directory` where the acto files exist (required if --file is not set)")
	file := flag.String("file", "", "A `file` that contains acto resources (required if --dir is not set)")
	recursive := flag.Bool("r", false, "Reads directory recursiveness (valid only if --dir is set)")
	help := flag.Bool("help", false, "Show this help output")
	version := flag.Bool("version", false, "Show the current version of acto")

	flag.Parse()

	return &Flags{
		Directory: *directory,
		File:      *file,
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

func (flags *Flags) GetUsage() {
	flag.Usage()
}

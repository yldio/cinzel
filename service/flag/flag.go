// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package flag

import "flag"

type Flags struct {
	Directory string
	File      string
	Recursive bool
}

func NewParseFlags() *Flags {
	directory := flag.String("dir", "", "A `directory` where the atos files exist (sub-directories included) (required if --file is not set)")
	file := flag.String("file", "", "A `file` that contains atos resources (required if --dir is not set)")
	recursive := flag.Bool("r", false, "Reads directory recursiveness (valid only if --dir is set)")

	flag.Parse()

	return &Flags{
		Directory: *directory,
		File:      *file,
		Recursive: *recursive,
	}
}

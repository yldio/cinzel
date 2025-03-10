// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"io"
	"os"

	"github.com/yldio/cinzel/internal/command"
	"github.com/yldio/cinzel/provider"
	"github.com/yldio/cinzel/provider/github"
)

var (
	version = "unknown"
)

func run(writer io.Writer, v string) error {
	cmd := command.New(writer, v)

	providers := []provider.Provider{
		github.New(),
	}

	return cmd.Execute(os.Args, providers)
}

func main() {
	if err := run(os.Stdout, version); err != nil {
		os.Exit(1)
	}
}

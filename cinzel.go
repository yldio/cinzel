// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"io"
	"os"

	"github.com/yldio/cinzel/internal/command"
	"github.com/yldio/cinzel/provider"
	"github.com/yldio/cinzel/provider/github"
	"github.com/yldio/cinzel/provider/gitlab"
)

var (
	version = "unknown"
)

func run(writer io.Writer, v string) error {
	cmd := command.New(writer, v)

	providers := []provider.Provider{
		github.New(),
		gitlab.New(),
	}

	return cmd.Execute(os.Args, providers)
}

func main() {
	if err := run(os.Stdout, version); err != nil {
		// A diff that could not reach an answer exits 2, so a workflow can
		// tell the generated files having drifted from cinzel not having
		// run. Every other failure is 1, the way it has always been.
		if errors.Is(err, command.ErrCannotRun) {
			os.Exit(2)
		}

		os.Exit(1)
	}
}

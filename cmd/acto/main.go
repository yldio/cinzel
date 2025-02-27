// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

// Package acto
package main

import (
	"context"
	"log"
	"os"

	"github.com/yldio/acto/command"
	"github.com/yldio/acto/provider/github"
)

var (
	version = ""
)

func main() {
	cmd := command.New(version)

	gh := github.New()
	cmd.AddCommand(gh)

	if err := cmd.Cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal("err", err)
	}
}

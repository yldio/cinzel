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
	version = "local"
)

func main() {
	// createBuildInfo()

	cmd := command.New(version)

	gh := github.New()
	cmd.AddCommand(gh)

	if err := cmd.Cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal("err", err)
	}
}

// func createBuildInfo() {
// 	buildInfo, ok := debug.ReadBuildInfo()
// 	if !ok {
// 		fmt.Println("not ok")
// 	}

// 	fmt.Println(buildInfo.Main)
// 	fmt.Println(buildInfo)
// }

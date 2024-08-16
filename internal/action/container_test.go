// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestContainer(t *testing.T) {
	type Test struct {
		name   string
		have   *ContainerConfig
		expect Container
	}

	var have_1 = ContainerConfig{
		Image: "node:18",
		Credentials: CredentialsConfig{
			Username: "${{ github.actor }}",
			Password: "${{ secrets.github_token }}",
		},
		Env: EnvConfig{
			Variable: []VariableConfig{
				{
					Name:  "NODE_ENV",
					Value: cty.StringVal("development"),
				},
			},
		},
		Ports: []int16{80},
		Volumes: []string{
			"my_docker_volume:/volume_mount",
			"/data/my_data",
			"/source/directory:/destination/directory",
		},
		Options: "--cpus 1",
	}
	var expect_1 = Container{
		Image: "node:18",
		Credentials: Credentials{
			Username: "${{ github.actor }}",
			Password: "${{ secrets.github_token }}",
		},
		Env: Env{
			"NODE_ENV": "development",
		},
		Ports: []int16{80},
		Volumes: []string{
			"my_docker_volume:/volume_mount",
			"/data/my_data",
			"/source/directory:/destination/directory",
		},
		Options: "--cpus 1",
	}

	var tests = []Test{
		{"with defined container", &have_1, expect_1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}
}

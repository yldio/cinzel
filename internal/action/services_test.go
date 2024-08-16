// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestServices(t *testing.T) {
	type Test struct {
		name   string
		have   *ServicesConfig
		expect Services
	}

	name_1 := "nginx"
	image_1 := "nginx"
	port_1_1 := "8080:80"
	ports_1 := []string{port_1_1}
	volume_1_1 := "my_docker_volume:/volume_mount"
	volume_1_2 := "/data/my_data"
	volume_1_3 := "/source/directory:/destination/directory"
	volumes_1 := []string{
		volume_1_1,
		volume_1_2,
		volume_1_3,
	}
	name_2 := "redis"
	image_2 := "redis"
	port_2_1 := "6379/tcp"
	ports_2 := []string{port_2_1}
	options_2 := "--cpus 1"

	var have_1 = ServicesConfig{
		{
			Name:  name_1,
			Image: &image_1,
			Ports: &ports_1,
			Credentials: &CredentialsConfig{
				Username: "${{ github.actor }}",
				Password: "${{ secrets.github_token }}",
			},
			Volumes: &volumes_1,
		},
		{
			Name:  name_2,
			Image: &image_2,
			Env: &EnvConfig{
				Variable: []VariableConfig{
					{
						Name:  "NODE_ENV",
						Value: cty.StringVal("development"),
					},
				},
			},
			Ports:   &ports_2,
			Options: &options_2,
		},
	}
	var expect_1 = Services{
		"nginx": Service{
			Name:  "nginx",
			Image: "nginx",
			Credentials: Credentials{
				Username: "${{ github.actor }}",
				Password: "${{ secrets.github_token }}",
			},
			Ports: []string{"8080:80"},
			Volumes: []string{
				"my_docker_volume:/volume_mount",
				"/data/my_data",
				"/source/directory:/destination/directory",
			},
		},
		"redis": Service{
			Name:  "redis",
			Image: "redis",
			Env: map[string]any{
				"NODE_ENV": "development",
			},
			Ports:   []string{"6379/tcp"},
			Options: "--cpus 1",
		},
	}

	var tests = []Test{
		{"with defined services", &have_1, expect_1},
		{"without service", nil, nil},
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

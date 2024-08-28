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

	name1 := "nginx"
	image1 := "nginx"
	port11 := "8080:80"
	ports1 := []string{port11}
	volume11 := "my_docker_volume:/volume_mount"
	volume12 := "/data/my_data"
	volume13 := "/source/directory:/destination/directory"
	volumes1 := []string{
		volume11,
		volume12,
		volume13,
	}
	name2 := "redis"
	image2 := "redis"
	port21 := "6379/tcp"
	ports2 := []string{port21}
	options2 := "--cpus 1"

	var have1 = ServicesConfig{
		{
			Name:  name1,
			Image: &image1,
			Ports: &ports1,
			Credentials: &CredentialsConfig{
				Username: "${{ github.actor }}",
				Password: "${{ secrets.github_token }}",
			},
			Volumes: &volumes1,
		},
		{
			Name:  name2,
			Image: &image2,
			Env: &EnvConfig{
				Variable: []VariableConfig{
					{
						Name:  "NODE_ENV",
						Value: cty.StringVal("development"),
					},
				},
			},
			Ports:   &ports2,
			Options: &options2,
		},
	}
	var expect1 = Services{
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
		{"with defined services", &have1, expect1},
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

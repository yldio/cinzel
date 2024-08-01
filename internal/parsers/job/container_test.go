// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
	"github.com/zclconf/go-cty/cty"
)

func TestJobOnlyWithContainer(t *testing.T) {

	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  container {
    image = "node:18"

    credentials {
      username = "$${{ github.actor }}"
      password = "$${{ secrets.github_token }}"
    }

    env {
      variable {
        name = "NODE_ENV"
        value = "development"
      }
    }

    volumes = [
      "my_docker_volume:/volume_mount",
      "/data/my_data",
      "/source/directory:/destination/directory",
    ]
    
    ports = [80]

    options = "--cpus 1"
  }
}
`

		var got_hcl HclConfig

		if err := parsers.HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Container: ContainerConfig{
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
					},
				},
			},
		}

		if !reflect.DeepEqual(got_hcl, expected_hcl) {
			t.FailNow()
		}

		got_parsed, err := got_hcl.Parse()
		if err != nil {
			t.FailNow()
		}

		expected_parsed := Jobs{
			"job_1": Job{
				Id: "job_1",
				Container: Container{
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
				},
			},
		}

		if !reflect.DeepEqual(got_parsed, expected_parsed) {
			t.FailNow()
		}

		got_yaml, err := parsers.Convert(got_parsed)
		if err != nil {
			t.FailNow()
		}

		expected_yaml := `job_1:
  container:
    image: node:18
    credentials:
      username: ${{ github.actor }}
      password: ${{ secrets.github_token }}
    env:
      NODE_ENV: development
    ports:
    - 80
    volumes:
    - my_docker_volume:/volume_mount
    - /data/my_data
    - /source/directory:/destination/directory
    options: --cpus 1
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}

// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
	"github.com/zclconf/go-cty/cty"
)

func TestJobOnlyPropServices(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  service "nginx" {
    image = "nginx"
    ports = ["8080:80"]
    credentials {
      username = "$${{ github.actor }}"
      password = "$${{ secrets.github_token }}"
    }

    volumes = [
      "my_docker_volume:/volume_mount",
      "/data/my_data",
      "/source/directory:/destination/directory",
    ]
  }

  service "redis" {
    image = "redis"
    ports = ["6379/tcp"]
    env {
      variable {
        name = "NODE_ENV"
        value = "development"
      }
    }
    options = "--cpus 1"
  }
}
`

		var got_hcl HclConfig

		if err := HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.Fail()
		}

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Services: ServicesConfig{
						{
							Name:  "nginx",
							Image: "nginx",
							Ports: []string{"8080:80"},
							Credentials: ServiceCredentialsConfig{
								Username: "${{ github.actor }}",
								Password: "${{ secrets.github_token }}",
							},
							Volumes: []string{
								"my_docker_volume:/volume_mount",
								"/data/my_data",
								"/source/directory:/destination/directory",
							},
						},
						{
							Name:  "redis",
							Image: "redis",
							Ports: []string{"6379/tcp"},
							Env: ServiceEnvConfig{
								Variable: []ServiceVariableConfig{
									{
										Name:  "NODE_ENV",
										Value: cty.StringVal("development"),
									},
								},
							},
							Options: "--cpus 1",
						},
					},
				},
			},
		}

		if !reflect.DeepEqual(got_hcl, expected_hcl) {
			t.Fail()
		}

		got_parsed, err := got_hcl.Parse()
		if err != nil {
			t.FailNow()
		}

		expected_parsed := Jobs{
			"job_1": Job{
				Id: "job_1",
				Services: Services{
					"nginx": Service{
						Name:  "nginx",
						Image: "nginx",
						Credentials: ServiceCredentials{
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
				},
			},
		}

		if !reflect.DeepEqual(got_parsed, expected_parsed) {
			t.FailNow()
		}

		got_yaml, err := parsers.Convert(got_parsed)
		if err != nil {
			t.Fail()
		}

		expected_yaml := `job_1:
  services:
    nginx:
      image: nginx
      credentials:
        username: ${{ github.actor }}
        password: ${{ secrets.github_token }}
      ports:
      - "8080:80"
      volumes:
      - my_docker_volume:/volume_mount
      - /data/my_data
      - /source/directory:/destination/directory
    redis:
      image: redis
      env:
        NODE_ENV: development
      ports:
      - 6379/tcp
      options: --cpus 1
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.Fail()
		}
	})
}

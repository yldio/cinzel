// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJob(t *testing.T) {

	t.Run("convert from hcl: job", func(t *testing.T) {
		have_hcl := `job "job_1" {}

job "job_2" {
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
}
`

		var got_hcl HclConfig

		if err := HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.Fail()
		}

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id:       "job_1",
					Services: nil,
				},
				{
					Id: "job_2",
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
			},
			"job_2": Job{
				Id: "job_2",
				Services: Services{
					"nginx": Service{
						Name:  "nginx",
						Image: "nginx",
						Credentials: ServiceCredentials{
							Username: "${{ github.actor }}",
							Password: "${{ secrets.github_token }}",
						},
						Env:   nil,
						Ports: []string{"8080:80"},
						Volumes: []string{
							"my_docker_volume:/volume_mount",
							"/data/my_data",
							"/source/directory:/destination/directory",
						},
						Options: "",
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

		expected_yaml := `job_1: {}
job_2:
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
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}

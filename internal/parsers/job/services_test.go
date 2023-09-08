package job

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestParseServices(t *testing.T) {
	ports_1 := []string{"8080:80"}
	ports_2 := []string{"6379/tcp"}
	volumes := []string{
		"my_docker_volume:/volume_mount",
		"/data/my_data",
		"/source/directory:/destination/directory",
	}

	t.Run("convert from hcl: services", func(t *testing.T) {
		have := []byte(`job {
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
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Services ServicesConfig `hcl:"service,block"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Jobs[0].Services

		expected := ServicesConfig{
			{
				Name:  "nginx",
				Image: "nginx",
				Ports: ports_1,
				Credentials: ServiceCredentialsConfig{
					Username: "${{ github.actor }}",
					Password: "${{ secrets.github_token }}",
				},
				Volumes: volumes,
			},
			{
				Name:  "redis",
				Image: "redis",
				Ports: ports_2,
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
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: services", func(t *testing.T) {
		have := ServicesConfig{
			{
				Name:  "nginx",
				Image: "nginx",
				Ports: ports_1,
				Credentials: ServiceCredentialsConfig{
					Username: "${{ github.actor }}",
					Password: "${{ secrets.github_token }}",
				},
				Volumes: volumes,
			},
			{
				Name:  "redis",
				Image: "redis",
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
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := Services{
			"nginx": Service{
				Name:  "nginx",
				Image: "nginx",
				Credentials: ServiceCredentials{
					Username: "${{ github.actor }}",
					Password: "${{ secrets.github_token }}",
				},
				Ports:   ports_1,
				Volumes: volumes,
			},
			"redis": Service{
				Name:  "redis",
				Image: "redis",
				Env: map[string]any{
					"NODE_ENV": "development",
				},
				Options: "--cpus 1",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: services", func(t *testing.T) {
		have := TestingServices{
			Services{
				"nginx": Service{
					Name:  "nginx",
					Image: "nginx",
					Credentials: ServiceCredentials{
						Username: "${{ github.actor }}",
						Password: "${{ secrets.github_token }}",
					},
					Ports:   ports_1,
					Volumes: volumes,
				},
				"redis": Service{
					Name:  "redis",
					Image: "redis",
					Env: map[string]any{
						"NODE_ENV": "development",
					},
					Options: "--cpus 1",
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`services:
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
    options: --cpus 1
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}

package job

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestParseContainer(t *testing.T) {

	image := "node:18"
	port := int16(80)
	ports := []int16{80}
	volumes := []string{
		"my_docker_volume:/volume_mount",
		"/data/my_data",
		"/source/directory:/destination/directory",
	}
	options := "--cpus 1"

	t.Run("convert from hcl: container", func(t *testing.T) {
		have := []byte(`job {
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
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Container *ContainerConfig `hcl:"container,block"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Jobs[0].Container

		expected := ContainerConfig{
			Image: &image,
			Credentials: &CredentialsConfig{
				Username: "${{ github.actor }}",
				Password: "${{ secrets.github_token }}",
			},
			Env: &EnvConfig{
				Variable: []VariableConfig{
					{
						Name:  "NODE_ENV",
						Value: cty.StringVal("development"),
					},
				},
			},
			Ports:   &ports,
			Volumes: &volumes,
			Options: &options,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: container", func(t *testing.T) {
		have := ContainerConfig{
			Image: &image,
			Credentials: &CredentialsConfig{
				Username: "${{ github.actor }}",
				Password: "${{ secrets.github_token }}",
			},
			Env: &EnvConfig{
				Variable: []VariableConfig{
					{
						Name:  "NODE_ENV",
						Value: cty.StringVal("development"),
					},
				},
			},
			Ports:   &ports,
			Volumes: &volumes,
			Options: &options,
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := Container{
			Image: image,
			Credentials: Credentials{
				Username: "${{ github.actor }}",
				Password: "${{ secrets.github_token }}",
			},
			Env: Env{
				"NODE_ENV": "development",
			},
			Ports: []int16{
				port,
			},
			Volumes: []string{
				"my_docker_volume:/volume_mount",
				"/data/my_data",
				"/source/directory:/destination/directory",
			},
			Options: options,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: container", func(t *testing.T) {
		have := TestingContainer{
			Container{
				Image: image,
				Credentials: Credentials{
					Username: "${{ github.actor }}",
					Password: "${{ secrets.github_token }}",
				},
				Env: Env{
					"NODE_ENV": "development",
				},
				Ports: []int16{
					port,
				},
				Volumes: []string{
					"my_docker_volume:/volume_mount",
					"/data/my_data",
					"/source/directory:/destination/directory",
				},
				Options: options,
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`container:
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
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}

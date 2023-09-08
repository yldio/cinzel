package workflow

import (
	"reflect"
	"testing"
)

func TestDefaults(t *testing.T) {
	t.Run("convert from hcl: defaults", func(t *testing.T) {
		have := []byte(`workflow {
  defaults {
    run {
      shell = "bash"
      working_directory = "./scripts"
    }
  }
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				Defaults DefaultsConfig `hcl:"defaults,block"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Workflows[0].Defaults

		expected := DefaultsConfig{
			Run: RunConfig{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: env", func(t *testing.T) {
		have := DefaultsConfig{
			Run: RunConfig{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := DefaultsConfig{
			Run: RunConfig{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: permissions", func(t *testing.T) {
		have := TestingDefaults{
			Defaults: DefaultsConfig{
				Run: RunConfig{
					Shell:            "bash",
					WorkingDirectory: "./scripts",
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`defaults:
  run:
    shell: bash
    working-directory: ./scripts
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}

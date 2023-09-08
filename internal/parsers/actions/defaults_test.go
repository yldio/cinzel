package actions

import (
	"reflect"
	"testing"
)

func TestDefaults(t *testing.T) {
	t.Run("converts HCL to a struct", func(t *testing.T) {
		have := DefaultsConfig{
			Run: &RunConfig{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		got, err := have.ConvertFromHcl()
		if err != nil {
			t.Errorf(err.Error())
		}

		expected := Defaults{
			Run: Run{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})

	t.Run("converts struct to Yaml", func(t *testing.T) {
		have := Defaults{
			Run: Run{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		got, err := have.ConvertToYaml()
		if err != nil {
			t.Errorf(err.Error())
		}

		expected := DefaultsYaml{
			Run: RunYaml{
				Shell:            "bash",
				WorkingDirectory: "./scripts",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})
}

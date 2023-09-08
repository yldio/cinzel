package actions

import (
	"reflect"
	"testing"
)

func TestEnv(t *testing.T) {
	t.Run("converts HCL to a struct", func(t *testing.T) {
		have := EnvsConfig{
			Name:  "Env1",
			Value: "Env 1",
		}

		got, err := have.ConvertFromHcl()
		if err != nil {
			t.Errorf(err.Error())
		}

		expected := Env{
			Name:  "Env1",
			Value: "Env 1",
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})

	t.Run("converts struct to Yaml", func(t *testing.T) {
		have := Envs{{
			Name:  "Env1",
			Value: "Env 1",
		}}

		got, err := have.ConvertToYaml()
		if err != nil {
			t.Errorf(err.Error())
		}

		expected := EnvsYaml{
			"Env1": "Env 1",
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})
}

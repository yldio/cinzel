package actions

import (
	"reflect"
	"testing"
)

func TestPermissions(t *testing.T) {
	t.Run("converts HCL to a struct", func(t *testing.T) {
		actionRead := Read
		actionWrite := Write
		actionNone := None
		have := PermissionsConfig{
			Actions:      &actionRead,
			Attestations: &actionRead,
			Checks:       &actionWrite,
			Contents:     &actionNone,
		}

		expected := Permissions{
			Permission{Perm: "Actions", Value: "read"},
			Permission{Perm: "Attestations", Value: "read"},
			Permission{Perm: "Checks", Value: "write"},
			Permission{Perm: "Contents", Value: "none"},
		}

		got, err := have.ConvertFromHcl()
		if err != nil {
			t.Errorf(err.Error())
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})

	t.Run("converts struct to Yaml", func(t *testing.T) {
		have := Permissions{
			Permission{Perm: "Actions", Value: "read"},
		}

		expected := PermissionsYaml{
			Actions: "read",
		}

		got, err := have.ConvertToYaml()
		if err != nil {
			t.Errorf(err.Error())
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})
}

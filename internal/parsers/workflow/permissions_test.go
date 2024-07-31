package workflow

import (
	"reflect"
	"testing"
)

func TestPermissions(t *testing.T) {
	t.Run("convert from hcl: permissions", func(t *testing.T) {
		have := []byte(`workflow {
  permissions {
    actions      = "read"
    attestations = "write"
    checks       = "none"
  }
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				Permissions *PermissionsConfig `hcl:"permissions,block"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Workflows[0].Permissions

		actions := Read
		attestations := Write
		checks := None

		expected := PermissionsConfig{
			Actions:      &actions,
			Attestations: &attestations,
			Checks:       &checks,
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: permissions", func(t *testing.T) {
		actions := Read
		attestations := Write
		checks := None

		have := PermissionsConfig{
			Actions:      &actions,
			Attestations: &attestations,
			Checks:       &checks,
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := Permissions{
			Actions:      &actions,
			Attestations: &attestations,
			Checks:       &checks,
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: permissions", func(t *testing.T) {
		actions := Read
		attestations := Write
		checks := None

		have := TestingPermissions{
			Permissions{
				Actions:      &actions,
				Attestations: &attestations,
				Checks:       &checks,
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`permissions:
  actions: read
  attestations: write
  checks: none
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}

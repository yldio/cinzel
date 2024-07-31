package step

import (
	"reflect"
	"testing"
)

func TestParseShell(t *testing.T) {
	t.Run("convert from hcl: shell", func(t *testing.T) {
		have := []byte(`step {
  shell = "bash"
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Shell *ShellConfig `hcl:"shell,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Steps[0].Shell

		expected := ShellConfig("bash")

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: shell", func(t *testing.T) {
		have := ShellConfig("bash")

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := "bash"

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: shell", func(t *testing.T) {
		have := TestingShell{
			Shell: "bash",
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`shell: bash
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}

package step

import (
	"reflect"
	"testing"
)

func TestParseWorkingDirectory(t *testing.T) {
	t.Run("convert from hcl: working-directory", func(t *testing.T) {
		have := []byte(`step {
  working_directory = "./temp"
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				WorkingDirectory *WorkingDirectoryConfig `hcl:"working_directory,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Steps[0].WorkingDirectory

		expected := WorkingDirectoryConfig("./temp")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: working-directory", func(t *testing.T) {
		have := WorkingDirectoryConfig("./temp")

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := "./temp"

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: working-directory", func(t *testing.T) {
		have := TestingWorkingDirectory{
			WorkingDirectory: "./temp",
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`working-directory: ./temp
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}

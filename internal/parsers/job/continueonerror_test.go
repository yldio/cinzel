package job

import (
	"reflect"
	"testing"
)

func TestParseContinueOnError(t *testing.T) {
	t.Run("convert from hcl: continue-on-error", func(t *testing.T) {
		have := []byte(`job {
  continue_on_error = true
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				ContinueOnError *ContinueOnErrorConfig `hcl:"continue_on_error,attr"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Jobs[0].ContinueOnError

		expected := ContinueOnErrorConfig(true)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: continue-on-error", func(t *testing.T) {
		have := ContinueOnErrorConfig(true)

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := true

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: continue-on-error", func(t *testing.T) {
		have := TestingContinueOnError{
			ContinueOnError: true,
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`continue-on-error: true
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}

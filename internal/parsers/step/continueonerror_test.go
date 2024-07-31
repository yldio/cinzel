package step

import (
	"reflect"
	"testing"
)

func TestParseContinueOnError(t *testing.T) {
	t.Run("convert from hcl: continue-on-error", func(t *testing.T) {
		have := []byte(`step {
  continue_on_error = true
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				ContinueOnError *ContinueOnErrorConfig `hcl:"continue_on_error,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Steps[0].ContinueOnError

		expected := ContinueOnErrorConfig(true)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: continue-on-error", func(t *testing.T) {
		have := ContinueOnErrorConfig(true)

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := true

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: continue-on-error", func(t *testing.T) {
		have := TestingContinueOnError{
			ContinueOnError: true,
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`continue-on-error: true
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}

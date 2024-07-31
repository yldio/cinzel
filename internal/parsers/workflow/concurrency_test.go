package workflow

import (
	"reflect"
	"testing"
)

func TestConcurrency(t *testing.T) {
	t.Run("convert from hcl: concurrency", func(t *testing.T) {
		have := []byte(`workflow {
  concurrency {
    group = "$${{ github.workflow }}-$${{ github.ref }}"
    cancel_in_progress = true
  }
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				Concurrency ConcurrencyConfig `hcl:"concurrency,block"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := hclConfig.Workflows[0].Concurrency

		cancelInProgress := true

		expected := ConcurrencyConfig{
			Group:            "${{ github.workflow }}-${{ github.ref }}",
			CancelInProgress: &cancelInProgress,
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: concurrency", func(t *testing.T) {
		cancelInProgress := true

		have := ConcurrencyConfig{
			Group:            "${{ github.workflow }}-${{ github.ref }}",
			CancelInProgress: &cancelInProgress,
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := ConcurrencyConfig{
			Group:            "${{ github.workflow }}-${{ github.ref }}",
			CancelInProgress: &cancelInProgress,
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: concurrency", func(t *testing.T) {
		cancelInProgress := true

		have := TestingConcurrency{
			Concurrency: ConcurrencyConfig{
				Group:            "${{ github.workflow }}-${{ github.ref }}",
				CancelInProgress: &cancelInProgress,
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}

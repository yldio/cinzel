package actions

import (
	"reflect"
	"testing"
)

func TestConcurrency(t *testing.T) {
	t.Run("converts HCL to a struct", func(t *testing.T) {
		have := ConcurrencyConfig{
			Group:            "${{ github.head_ref || github.run_id }}",
			CancelInProgress: true,
		}

		got, err := have.ConvertFromHcl()
		if err != nil {
			t.Errorf(err.Error())
		}

		expected := Concurrency{
			Group:            "${{ github.head_ref || github.run_id }}",
			CancelInProgress: true,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})

	t.Run("converts struct to Yaml", func(t *testing.T) {
		have := Concurrency{
			Group:            "${{ github.head_ref || github.run_id }}",
			CancelInProgress: true,
		}

		got, err := have.ConvertToYaml()
		if err != nil {
			t.Errorf(err.Error())
		}

		expected := ConcurrencyYaml{
			Group:            "${{ github.head_ref || github.run_id }}",
			CancelInProgress: true,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf(err.Error())
		}
	})
}

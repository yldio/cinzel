package actions

import (
	"testing"
)

func TestEvent(t *testing.T) {
	t.Run("converts HCL to a struct", func(t *testing.T) {
		// onConfig := OnConfig("push")
		// have_single := EventConfig{
		// 	On: &onConfig,
		// }

		// onAsListConfig := OnAsListConfig{"push", "pull-request"}
		// have_multi := EventConfig{
		// 	OnAsList: &onAsListConfig,
		// }

		// have_struct := EventConfig{
		// 	OnByFilter: []*OnByFilterConfig{{
		// 		Event:  "label",
		// 		Filter: "types",
		// 		Values: []string{"created"},
		// 	}},
		// }

		// expected_single := Event{
		// 	On: "push",
		// }

		// expected_multi := Event{
		// 	OnAsList: []string{"push", "pull-request"},
		// }

		// expected_struct := Event{
		// 	OnByFilter: []OnByFilter{{
		// 		Event:  "label",
		// 		Filter: "types",
		// 		Values: []string{"created"},
		// 	}},
		// }

		// got_single, err := have_single.ConvertFromHcl()
		// if err != nil {
		// 	t.Errorf(err.Error())
		// }

		// got_multi, err := have_multi.ConvertFromHcl()
		// if err != nil {
		// 	t.Errorf(err.Error())
		// }

		// got_struct, err := have_struct.ConvertFromHcl()
		// if err != nil {
		// 	t.Errorf(err.Error())
		// }

		// if !reflect.DeepEqual(got_single, expected_single) {
		// 	t.Errorf(err.Error())
		// }

		// if !reflect.DeepEqual(got_multi, expected_multi) {
		// 	t.Errorf(err.Error())
		// }

		// if !reflect.DeepEqual(got_struct, expected_struct) {
		// 	t.Errorf(err.Error())
		// }
	})

	t.Run("converts struct to Yaml", func(t *testing.T) {
		// have_single := Event{
		// 	On: "push",
		// }

		// have_multi := Event{
		// 	OnAsList: []string{"push", "pull-request"},
		// }

		// got_single, err := have_single.ConvertToYaml()
		// if err != nil {
		// 	t.Errorf(err.Error())
		// }

		// got_multi, err := have_multi.ConvertToYaml()
		// if err != nil {
		// 	t.Errorf(err.Error())
		// }

		// expected_single := "push"

		// expected_multi := []string{"push", "pull-request"}

		// if !reflect.DeepEqual(got_single, expected_single) {
		// 	t.Errorf(err.Error())
		// }

		// if !reflect.DeepEqual(got_multi, expected_multi) {
		// 	t.Errorf(err.Error())
		// }
	})
}

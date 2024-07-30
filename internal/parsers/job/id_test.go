// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

type MockJobsIdConfig struct {
	Jobs []IdConfig `hcl:"job,block"`
}

func (config *MockJobsIdConfig) Parse() (Jobs, error) {
	jobs := make(Jobs)
	for _, job := range config.Jobs {
		parsedJob, err := job.Parse()
		if err != nil {
			return Jobs{}, nil
		}

		jobs[parsedJob.Id] = parsedJob
	}
	return jobs, nil
}

func TestJobOnlyPropId(t *testing.T) {

	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {}
`

		var got_hcl MockJobsIdConfig

		if err := HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.Fail()
		}

		expected_hcl := MockJobsIdConfig{
			Jobs: []IdConfig{
				{
					Id: "job_1",
				},
			},
		}

		if !reflect.DeepEqual(got_hcl, expected_hcl) {
			t.Fail()
		}

		got_parsed, err := got_hcl.Parse()
		if err != nil {
			t.Fail()
		}

		expected_parsed := Jobs{
			"job_1": Job{
				Id: "job_1",
			},
		}

		if !reflect.DeepEqual(got_parsed, expected_parsed) {
			t.Fail()
		}

		got_yaml, err := parsers.Convert(got_parsed)
		if err != nil {
			t.Fail()
		}

		expected_yaml := `job_1: {}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}

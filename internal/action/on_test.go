// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestOn(t *testing.T) {
	type Test struct {
		name   string
		have   *OnConfig
		expect On
	}

	var push = TriggerPush.ToString()
	var pullRequest = TriggerPullRequest.ToString()
	var singleEvent = cty.StringVal(push)
	var multipleEvent = cty.TupleVal([]cty.Value{
		cty.StringVal(push),
		cty.StringVal(pullRequest),
	})
	var eventPush = EventConfig{
		Event:    push,
		Branches: &[]string{"main", "mona/octocat", "releases/**"},
		Tags:     &[]string{"v2", "v1.*"},
	}
	var eventPageBuild = EventConfig{
		Event: TriggerPageBuild.ToString(),
	}
	var pushEvent = []*EventConfig{
		&eventPush,
		&eventPageBuild,
	}
	var pullRequestEvent = []*EventConfig{
		{
			Event:    pullRequest,
			Branches: &[]string{"main", "mona/octocat", "releases/**"},
		},
	}

	var pullRequestActivity = []*ActivityConfig{
		{
			Id:    TriggerLabel.ToString(),
			Types: []string{ActivityCreated.ToString()},
		},
	}

	var have1 = OnConfig{
		Events: &singleEvent,
	}
	var expect1 = push

	var have2 = OnConfig{
		Events: &multipleEvent,
	}
	var expect2 = []string{push, pullRequest}

	var have3 = OnConfig{
		Event: pushEvent,
	}
	var expect3 = map[string]map[string][]string{
		"push": {
			"tags":     {"v2", "v1.*"},
			"branches": {"main", "mona/octocat", "releases/**"},
		},
		"page_build": {},
	}

	var have4 = OnConfig{
		Event:    pullRequestEvent,
		Activity: pullRequestActivity,
	}
	var expect4 = map[string]map[string][]string{
		"pull_request": {
			"branches": {"main", "mona/octocat", "releases/**"},
		},
		"label": {
			"types": {"created"},
		},
	}

	var tests = []Test{
		{"with defined single event", &have1, expect1},
		{"with defined multiple events", &have2, expect2},
		{"with defined push event with tags and branches and defined page_build", &have3, expect3},
		{"with defined pull-request event with branches", &have4, expect4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}
}

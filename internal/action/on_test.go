// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

// import (
// 	"reflect"
// 	"testing"

// 	"github.com/zclconf/go-cty/cty"
// )

// func TestOn(t *testing.T) {
// 	type Test struct {
// 		name   string
// 		have   *OnConfig
// 		expect On
// 	}

// 	var push = "push"
// 	var pullRequest = "pull-request"
// 	var singleEvent = cty.StringVal(push)
// 	var multipleEvent = cty.TupleVal([]cty.Value{
// 		cty.StringVal(push),
// 		cty.StringVal(pullRequest),
// 	})
// 	var pushEvent = []EventConfig{
// 		{
// 			Event:    push,
// 			Branches: []string{"main", "mona/octocat", "releases/**"},
// 			Tags:     []string{"v2", "v1.*"},
// 		},
// 		{
// 			Event: "page_build",
// 		},
// 	}
// 	var pullRequestEvent = []EventConfig{
// 		{
// 			Event:    pullRequest,
// 			Branches: []string{"main", "mona/octocat", "releases/**"},
// 		},
// 	}

// 	var pullRequestActivity = []ActivityConfig{
// 		{
// 			Activity: "label",
// 			Types:    []string{"created"},
// 		},
// 	}

// 	var have_1 = OnConfig{
// 		Events: &singleEvent,
// 	}
// 	var expect_1 = push

// 	var have_2 = OnConfig{
// 		Events: &multipleEvent,
// 	}
// 	var expect_2 = []string{push, pullRequest}

// 	var have_3 = OnConfig{
// 		Event: &pushEvent,
// 	}
// 	var expect_3 = map[string]map[string][]string{
// 		"push": {
// 			"tags":     {"v2", "v1.*"},
// 			"branches": {"main", "mona/octocat", "releases/**"},
// 		},
// 		"page_build": {},
// 	}

// 	var have_4 = OnConfig{
// 		Event:    &pullRequestEvent,
// 		Activity: &pullRequestActivity,
// 	}
// 	var expect_4 = map[string]map[string][]string{
// 		"pull-request": {
// 			"branches": {"main", "mona/octocat", "releases/**"},
// 		},
// 		"label": {
// 			"types": {"created"},
// 		},
// 	}

// 	var tests = []Test{
// 		{"with defined single event", &have_1, expect_1},
// 		{"with defined multiple events", &have_2, expect_2},
// 		{"with defined push event with tags and branches and defined page_build", &have_3, expect_3},
// 		{"with defined pull-request event with branches", &have_4, expect_4},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.have.Parse()
// 			if err != nil {
// 				t.Error(err.Error())
// 			}

// 			if !reflect.DeepEqual(got, tt.expect) {
// 				t.Fatalf("%s - failed", tt.name)
// 			}
// 		})
// 	}
// }

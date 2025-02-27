// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestOn(t *testing.T) {
	// type Test struct {
	// 	name         string
	// 	have         EventsConfig
	// 	expect       On
	// 	errorMessage error
	// }

	// var filename = "dummy-file"

	// var eventPush = "push"
	// // var eventPullRequest = "pull_request"
	// // var eventLabel = "label"
	// // var eventSchedule = "schedule"
	// // var eventWorkflowCall = "workflow_call"
	// // var eventWorkflowRun = "workflow_run"
	// // var eventWorkflowDispatch = "workflow_dispatch"

	// var branchMain = "main"
	// var branchReleases = "relrases/**"

	// branchesExp1, _ := hclsyntax.ParseExpression([]byte(fmt.Sprintf(`["%s", "%s"]`, branchMain, branchReleases)), filename, hcl.Pos{})

	// // 	var activityAsString = "label"
	// // 	var eventsAsListOfStrings = []string{"push", "pull_request"}
	// // 	var eventsWithBranches = []string{"main", "mona/octocat", "releases/**"}
	// // 	var eventsWithPaths = []string{"**.js"}
	// // 	var eventsWithTags = []string{"v2", "v1.*"}
	// // 	var eventsWithActivity = []string{ActivityCreated.ToString()}
	// // 	var eventsWithCron = []string{"cron: '30 5 * * 1,3'", "cron: '30 5 * * 2,4'"}

	// var haveEventPushWithBranches = EventConfig{
	// 	Identifier: eventPush,
	// 	Branches:   branchesExp1,
	// }

	// var haveEventsPushWithBranches = EventsConfig{
	// 	&haveEventPushWithBranches,
	// }

	// var expectEventsPushWithBranches = On{
	// 	TriggerPush: Event{
	// 		Name:     eventPush,
	// 		Branches: []*string{&branchMain, &branchReleases},
	// 	},
	// }

	// var tests = []Test{
	// 	{"event push with branches", haveEventsPushWithBranches, expectEventsPushWithBranches, nil},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := tt.have.Parse()

	// 		if tt.errorMessage == nil && err != nil {
	// 			t.Fatal(err.Error())
	// 		}

	// 		if tt.errorMessage == nil && !reflect.DeepEqual(got, tt.expect) {
	// 			t.Fatal(tt.name)
	// 		}

	// 		if tt.errorMessage != nil && tt.errorMessage.Error() != err.Error() {
	// 			t.Fatal(err)
	// 		}

	// 	})
	// }
}

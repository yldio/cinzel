// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestEventTrigger(t *testing.T) {
	t.Run("validate against defined event triggers", func(t *testing.T) {
		var list = []string{
			"branch_protection_rule",
			"check_run",
			"check_suite",
			"create",
			"delete",
			"deployment",
			"deployment_status",
			"discussion",
			"discussion_comment",
			"fork",
			"gollum",
			"issue_comment",
			"issues",
			"label",
			"merge_group",
			"milestone",
			"page_build",
			"project",
			"project_card",
			"project_column",
			"public",
			"pull_request",
			"pull_request_comment",
			"pull_request_review",
			"pull_request_review_comment",
			"pull_request_target",
			"push",
			"registry_package",
			"release",
			"repository_dispatch",
			"schedule",
			"status",
			"watch",
			"workflow_call",
			"workflow_dispatch",
			"workflow_run",
		}

		for _, event := range list {
			if !ValidateEventTrigger(event) {
				t.Fatal()
			}
		}

		if ValidateEventTrigger("random") {
			t.Fatal()
		}
	})
}

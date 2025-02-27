// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestActivityType(t *testing.T) {
	t.Run("validate activity types for defined event triggers", func(t *testing.T) {
		if !ValidateActivityTypeForEventTrigger(TriggerPullRequest, ActivityAssigned) {
			t.Fatal()
		}

		if ValidateActivityTypeForEventTrigger(TriggerPullRequest, ActivityCreated) {
			t.Fatal()
		}
	})

	t.Run("validate against defined activity types", func(t *testing.T) {
		var list = []string{
			"types",
			"answered",
			"assigned",
			"auto_merge_disabled",
			"auto_merge_enabled",
			"category_changed",
			"checks_requested",
			"closed",
			"completed",
			"converted",
			"converted_to_draft",
			"created",
			"deleted",
			"demilestoned",
			"dequeued",
			"dismissed",
			"edited",
			"enqueued",
			"in_progress",
			"labeled",
			"locked",
			"milestoned",
			"moved",
			"opened",
			"pinned",
			"prereleased",
			"published",
			"ready_for_review",
			"released",
			"reopened",
			"requested",
			"requested_action",
			"rerequested",
			"review_requested",
			"review_request_removed",
			"started",
			"submitted",
			"synchronize",
			"transferred",
			"unanswered",
			"unassigned",
			"unlabeled",
			"unlocked",
			"unpinned",
			"unpublished",
			"updated",
		}

		for _, activity := range list {
			if !ValidateActivityType(activity) {
				t.Fatal()
			}
		}

		if ValidateActivityType("random") {
			t.Fatal()
		}
	})
}

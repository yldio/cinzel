// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type EventTrigger string

const (
	TriggerBranchProtectionRule     EventTrigger = "branch_protection_rule"
	TriggerCheckRun                 EventTrigger = "check_run"
	TriggerCheckSuite               EventTrigger = "check_suite"
	TriggerCreate                   EventTrigger = "create"
	TriggerDelete                   EventTrigger = "delete"
	TriggerDeployment               EventTrigger = "deployment"
	TriggerDeploymentStatus         EventTrigger = "deployment_status"
	TriggerDiscussion               EventTrigger = "discussion"
	TriggerDiscussionComment        EventTrigger = "discussion_comment"
	TriggerFork                     EventTrigger = "fork"
	TriggerGollum                   EventTrigger = "gollum"
	TriggerIssueComment             EventTrigger = "issue_comment"
	TriggerIssues                   EventTrigger = "issues"
	TriggerLabel                    EventTrigger = "label"
	TriggerMergeGroup               EventTrigger = "merge_group"
	TriggerMilestone                EventTrigger = "milestone"
	TriggerPageBuild                EventTrigger = "page_build"
	TriggerProject                  EventTrigger = "project"
	TriggerProjectCard              EventTrigger = "project_card"
	TriggerProjectColumn            EventTrigger = "project_column"
	TriggerPublic                   EventTrigger = "public"
	TriggerPullRequest              EventTrigger = "pull_request"
	TriggerPullRequestComment       EventTrigger = "pull_request_comment"
	TriggerPullRequestReview        EventTrigger = "pull_request_review"
	TriggerPullRequestReviewComment EventTrigger = "pull_request_review_comment"
	TriggerPullRequestTarget        EventTrigger = "pull_request_target"
	TriggerPush                     EventTrigger = "push"
	TriggerRegistryPackage          EventTrigger = "registry_package"
	TriggerRelease                  EventTrigger = "release"
	TriggerRepositoryDispatch       EventTrigger = "repository_dispatch"
	TriggerSchedule                 EventTrigger = "schedule"
	TriggerStatus                   EventTrigger = "status"
	TriggerWatch                    EventTrigger = "watch"
	TriggerWorkflowCall             EventTrigger = "workflow_call"
	TriggerWorkflowDispatch         EventTrigger = "workflow_dispatch"
	TriggerWorkflowRun              EventTrigger = "workflow_run"
)

func (eventTrigger EventTrigger) ToString() string {
	return string(eventTrigger)
}

func ValidateEventTrigger(eventTrigger string) bool {
	switch eventTrigger {
	case TriggerBranchProtectionRule.ToString():
		return true
	case TriggerCheckRun.ToString():
		return true
	case TriggerCheckSuite.ToString():
		return true
	case TriggerCreate.ToString():
		return true
	case TriggerDelete.ToString():
		return true
	case TriggerDeployment.ToString():
		return true
	case TriggerDeploymentStatus.ToString():
		return true
	case TriggerDiscussion.ToString():
		return true
	case TriggerDiscussionComment.ToString():
		return true
	case TriggerFork.ToString():
		return true
	case TriggerGollum.ToString():
		return true
	case TriggerIssueComment.ToString():
		return true
	case TriggerIssues.ToString():
		return true
	case TriggerLabel.ToString():
		return true
	case TriggerMergeGroup.ToString():
		return true
	case TriggerMilestone.ToString():
		return true
	case TriggerPageBuild.ToString():
		return true
	case TriggerProject.ToString():
		return true
	case TriggerProjectCard.ToString():
		return true
	case TriggerProjectColumn.ToString():
		return true
	case TriggerPublic.ToString():
		return true
	case TriggerPullRequest.ToString():
		return true
	case TriggerPullRequestComment.ToString():
		return true
	case TriggerPullRequestReview.ToString():
		return true
	case TriggerPullRequestReviewComment.ToString():
		return true
	case TriggerPullRequestTarget.ToString():
		return true
	case TriggerPush.ToString():
		return true
	case TriggerRegistryPackage.ToString():
		return true
	case TriggerRelease.ToString():
		return true
	case TriggerRepositoryDispatch.ToString():
		return true
	case TriggerSchedule.ToString():
		return true
	case TriggerStatus.ToString():
		return true
	case TriggerWatch.ToString():
		return true
	case TriggerWorkflowCall.ToString():
		return true
	case TriggerWorkflowDispatch.ToString():
		return true
	case TriggerWorkflowRun.ToString():
		return true
	default:
		return false
	}
}


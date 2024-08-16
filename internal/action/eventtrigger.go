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
	TriggerIssue_comment            EventTrigger = "issue_comment"
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

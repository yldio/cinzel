// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

type ActivityType string

const (
	ActivityTypes                ActivityType = "types"
	ActivityAnswered             ActivityType = "answered"
	ActivityAssigned             ActivityType = "assigned"
	ActivityAutoMergeDisabled    ActivityType = "auto_merge_disabled"
	ActivityAutoMergeEnabled     ActivityType = "auto_merge_enabled"
	ActivityCategoryChanged      ActivityType = "category_changed"
	ActivityChecksRequested      ActivityType = "checks_requested"
	ActivityClosed               ActivityType = "closed"
	ActivityCompleted            ActivityType = "completed"
	ActivityConverted            ActivityType = "converted"
	ActivityConvertedToDraft     ActivityType = "converted_to_draft"
	ActivityCreated              ActivityType = "created"
	ActivityDeleted              ActivityType = "deleted"
	ActivityDemilestoned         ActivityType = "demilestoned"
	ActivityDequeued             ActivityType = "dequeued"
	ActivityDismissed            ActivityType = "dismissed"
	ActivityEdited               ActivityType = "edited"
	ActivityEnqueued             ActivityType = "enqueued"
	ActivityInProgress           ActivityType = "in_progress"
	ActivityLabeled              ActivityType = "labeled"
	ActivityLocked               ActivityType = "locked"
	ActivityMilestoned           ActivityType = "milestoned"
	ActivityMoved                ActivityType = "moved"
	ActivityOpened               ActivityType = "opened"
	ActivityPinned               ActivityType = "pinned"
	ActivityPrereleased          ActivityType = "prereleased"
	ActivityPublished            ActivityType = "published"
	ActivityReadyForReview       ActivityType = "ready_for_review"
	ActivityReleased             ActivityType = "released"
	ActivityReopened             ActivityType = "reopened"
	ActivityRequested            ActivityType = "requested"
	ActivityRequestedAction      ActivityType = "requested_action"
	ActivityRerequested          ActivityType = "rerequested"
	ActivityReviewRequested      ActivityType = "review_requested"
	ActivityReviewRequestRemoved ActivityType = "review_request_removed"
	ActivityStarted              ActivityType = "started"
	ActivitySubmitted            ActivityType = "submitted"
	ActivitySynchronize          ActivityType = "synchronize"
	ActivityTransferred          ActivityType = "transferred"
	ActivityUnanswered           ActivityType = "unanswered"
	ActivityUnassigned           ActivityType = "unassigned"
	ActivityUnlabeled            ActivityType = "unlabeled"
	ActivityUnlocked             ActivityType = "unlocked"
	ActivityUnpinned             ActivityType = "unpinned"
	ActivityUnpublished          ActivityType = "unpublished"
	ActivityUpdated              ActivityType = "updated"
)

type ActivitiesType map[EventTrigger][]ActivityType

func NewActivitiesType() ActivitiesType {
	return ActivitiesType{
		TriggerBranchProtectionRule:     {ActivityCreated, ActivityEdited, ActivityDeleted},
		TriggerCheckRun:                 {ActivityCreated, ActivityRerequested, ActivityCompleted, ActivityRequestedAction},
		TriggerCheckSuite:               {ActivityCompleted},
		TriggerCreate:                   nil,
		TriggerDelete:                   nil,
		TriggerDeployment:               nil,
		TriggerDeploymentStatus:         nil,
		TriggerDiscussion:               {ActivityCreated, ActivityEdited, ActivityDeleted, ActivityTransferred, ActivityPinned, ActivityUnpinned, ActivityLabeled, ActivityUnlabeled, ActivityLocked, ActivityUnlocked, ActivityCategoryChanged, ActivityAnswered, ActivityUnanswered},
		TriggerDiscussionComment:        {ActivityCreated, ActivityEdited, ActivityDeleted},
		TriggerFork:                     nil,
		TriggerGollum:                   nil,
		TriggerIssueComment:             {ActivityCreated, ActivityEdited, ActivityDeleted},
		TriggerIssues:                   {ActivityOpened, ActivityEdited, ActivityDeleted, ActivityTransferred, ActivityPinned, ActivityUnpinned, ActivityClosed, ActivityReopened, ActivityAssigned, ActivityUnassigned, ActivityLabeled, ActivityUnlabeled, ActivityLocked, ActivityUnlocked, ActivityMilestoned, ActivityDemilestoned},
		TriggerLabel:                    {ActivityCreated, ActivityEdited, ActivityDeleted},
		TriggerMergeGroup:               {ActivityChecksRequested},
		TriggerMilestone:                {ActivityCreated, ActivityClosed, ActivityOpened, ActivityEdited, ActivityDeleted},
		TriggerPageBuild:                {ActivityCreated, ActivityClosed, ActivityReopened, ActivityEdited, ActivityDeleted},
		TriggerProject:                  {ActivityCreated, ActivityClosed, ActivityReopened, ActivityEdited, ActivityDeleted},
		TriggerProjectCard:              {ActivityCreated, ActivityMoved, ActivityConverted, ActivityEdited, ActivityDeleted},
		TriggerProjectColumn:            {ActivityCreated, ActivityUpdated, ActivityMoved, ActivityDeleted},
		TriggerPublic:                   nil,
		TriggerPullRequest:              {ActivityAssigned, ActivityUnassigned, ActivityLabeled, ActivityUnlabeled, ActivityOpened, ActivityEdited, ActivityClosed, ActivityReopened, ActivitySynchronize, ActivityConvertedToDraft, ActivityLocked, ActivityUnlocked, ActivityEnqueued, ActivityDequeued, ActivityMilestoned, ActivityDemilestoned, ActivityReadyForReview, ActivityReviewRequested, ActivityReviewRequestRemoved, ActivityAutoMergeEnabled, ActivityAutoMergeDisabled},
		TriggerPullRequestComment:       nil,
		TriggerPullRequestReview:        {ActivitySubmitted, ActivityEdited, ActivityDismissed},
		TriggerPullRequestReviewComment: {ActivityCreated, ActivityEdited, ActivityDeleted},
		TriggerPullRequestTarget:        {ActivityAssigned, ActivityUnassigned, ActivityLabeled, ActivityUnlabeled, ActivityOpened, ActivityEdited, ActivityClosed, ActivityReopened, ActivitySynchronize, ActivityConvertedToDraft, ActivityReadyForReview, ActivityLocked, ActivityUnlocked, ActivityReviewRequested, ActivityReviewRequestRemoved, ActivityAutoMergeEnabled, ActivityAutoMergeDisabled},
		TriggerPush:                     nil,
		TriggerRegistryPackage:          {ActivityPublished, ActivityUpdated},
		TriggerRelease:                  {ActivityPublished, ActivityUnpublished, ActivityCreated, ActivityEdited, ActivityDeleted, ActivityPrereleased, ActivityReleased},
		TriggerRepositoryDispatch:       {},
		TriggerSchedule:                 nil,
		TriggerStatus:                   nil,
		TriggerWatch:                    {ActivityStarted},
		TriggerWorkflowCall:             nil,
		TriggerWorkflowDispatch:         nil,
		TriggerWorkflowRun:              {ActivityCompleted, ActivityRequested, ActivityInProgress},
	}
}

func (activityType ActivityType) ToString() string {
	return string(activityType)
}

func ValidateActivityType(activityType string) bool {
	switch activityType {
	case ActivityTypes.ToString():
		return true
	case ActivityAnswered.ToString():
		return true
	case ActivityAssigned.ToString():
		return true
	case ActivityAutoMergeDisabled.ToString():
		return true
	case ActivityAutoMergeEnabled.ToString():
		return true
	case ActivityCategoryChanged.ToString():
		return true
	case ActivityChecksRequested.ToString():
		return true
	case ActivityClosed.ToString():
		return true
	case ActivityCompleted.ToString():
		return true
	case ActivityConverted.ToString():
		return true
	case ActivityConvertedToDraft.ToString():
		return true
	case ActivityCreated.ToString():
		return true
	case ActivityDeleted.ToString():
		return true
	case ActivityDemilestoned.ToString():
		return true
	case ActivityDequeued.ToString():
		return true
	case ActivityDismissed.ToString():
		return true
	case ActivityEdited.ToString():
		return true
	case ActivityEnqueued.ToString():
		return true
	case ActivityInProgress.ToString():
		return true
	case ActivityLabeled.ToString():
		return true
	case ActivityLocked.ToString():
		return true
	case ActivityMilestoned.ToString():
		return true
	case ActivityMoved.ToString():
		return true
	case ActivityOpened.ToString():
		return true
	case ActivityPinned.ToString():
		return true
	case ActivityPrereleased.ToString():
		return true
	case ActivityPublished.ToString():
		return true
	case ActivityReadyForReview.ToString():
		return true
	case ActivityReleased.ToString():
		return true
	case ActivityReopened.ToString():
		return true
	case ActivityRequested.ToString():
		return true
	case ActivityRequestedAction.ToString():
		return true
	case ActivityRerequested.ToString():
		return true
	case ActivityReviewRequested.ToString():
		return true
	case ActivityReviewRequestRemoved.ToString():
		return true
	case ActivityStarted.ToString():
		return true
	case ActivitySubmitted.ToString():
		return true
	case ActivitySynchronize.ToString():
		return true
	case ActivityTransferred.ToString():
		return true
	case ActivityUnanswered.ToString():
		return true
	case ActivityUnassigned.ToString():
		return true
	case ActivityUnlabeled.ToString():
		return true
	case ActivityUnlocked.ToString():
		return true
	case ActivityUnpinned.ToString():
		return true
	case ActivityUnpublished.ToString():
		return true
	case ActivityUpdated.ToString():
		return true
	default:
		return false
	}
}

func ValidateActivityTypeForEventTrigger(eventTrigger EventTrigger, activityType ActivityType) bool {
	allowedActivityType := false
	activitiesType := NewActivitiesType()

	for _, activity := range activitiesType[eventTrigger] {
		if activity == activityType {
			allowedActivityType = true
			break
		}
	}

	return allowedActivityType
}

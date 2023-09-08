package actions

import (
	"fmt"
)

type EventConfig struct {
	On         *OnConfig           `hcl:"on,attr"`
	OnAsList   *OnAsListConfig     `hcl:"on_as_list,attr"`
	OnByFilter []*OnByFilterConfig `hcl:"on_by_filter,block"`
}

type Oner interface {
	ConvertEventFromHcl() (Event, error)
}

type OnConfig string

func (on *OnConfig) ConvertEventFromHcl() (Event, error) {
	event := Event{
		On: string(*on),
	}

	return event, nil
}

type OnAsListConfig []string

func (on *OnAsListConfig) ConvertEventFromHcl() (Event, error) {
	// TODO: validate value is part of the EventTrigger
	event := Event{
		OnAsList: *on,
	}

	return event, nil
}

type OnByFilterConfig struct {
	Event  string    `hcl:"event,attr"`
	Filter *string   `hcl:"filter,attr"`
	Values *[]string `hcl:"values,attr"`
}

func (on *OnByFilterConfig) ConvertEventFromHcl() (OnByFilter, error) {
	o := OnByFilter{
		// Event: on.Event,
	}

	// if on.Filter != nil {
	// 	o.Filter = on.Filter

	// 	if on.Values == nil {
	// 		return OnByFilter{}, errors.New("if `filter` is set then `values` must be set as well")
	// 	}

	// 	for _, value := range *on.Values {
	// 		// TODO: validate value is part of the EventTrigger
	// 		o.Values = append(o.Values, &value)
	// 	}
	// }

	return o, nil
}

type OnByFiltersConfig []OnByFilterConfig

func (o *OnByFiltersConfig) ConvertEventFromHcl() (Event, error) {
	return Event{}, nil
}

// func convertEventFromHcl(on Oner) (Event, error) {
// 	switch v := on.(type) {
// 	case *OnConfig:
// 		return on.ConvertEventFromHcl()
// 	case *OnAsListConfig:
// 		return on.ConvertEventFromHcl()
// 	case *OnByFiltersConfig:
// 		return on.ConvertEventFromHcl()
// 	default:
// 		return Event{}, fmt.Errorf("undefined Event.On %s", v)
// 	}
// }

func (config *EventConfig) ConvertFromHcl() (Event, error) {
	var event Event

	// if config.On != nil {
	// 	evt, err := convertEventFromHcl(config.On)
	// 	if err != nil {
	// 		return Event{}, err
	// 	}
	// 	event = evt
	// } else if config.OnAsList != nil {
	// 	evt, err := convertEventFromHcl(config.OnAsList)
	// 	if err != nil {
	// 		return Event{}, err
	// 	}
	// 	event = evt
	// } else if config.OnByFilter != nil {
	// 	var ons []OnByFilter
	// 	for _, e := range config.OnByFilter {
	// 		on, err := e.ConvertEventFromHcl()
	// 		if err != nil {
	// 			return Event{}, err
	// 		}
	// 		ons = append(ons, on)
	// 	}
	// 	event = Event{
	// 		OnByFilter: ons,
	// 	}
	// }

	return event, nil
}

type OnByFilter struct {
	Event  string    `hcl:"event,attr"`
	Filter *string   `hcl:"filter,attr"`
	Values []*string `hcl:"values,attr"`
}

type Event struct {
	On         string
	OnAsList   []string
	OnByFilter []OnByFilter
}

func (event *Event) ConvertToYaml() (EventsYaml, error) {
	var evt any
	// if event.On != "" {
	// 	evt = event.On
	// } else if event.OnAsList != nil {
	// 	evt = event.OnAsList
	// } else if event.OnByFilter != nil {
	// 	evt = event.OnByFilter
	// }

	yaml := evt

	return yaml, nil
}

type EventsYaml any

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

func ValidateEventTrigger(trigger string) bool {
	switch trigger {
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
	case TriggerIssue_comment.ToString():
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
	}

	return false
}

func ValidateActivityType(activityType string) bool {
	switch activityType {
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
	}

	return false
}

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

var (
	ActivitiesType = map[EventTrigger][]ActivityType{
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
		TriggerIssue_comment:            {ActivityCreated, ActivityEdited, ActivityDeleted},
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
)

func (activityType ActivityType) ToString() string {
	return string(activityType)
}

func CheckValid(s string) {
	for _, at := range ActivitiesType[TriggerWorkflowRun] {
		fmt.Println(at)
	}
}

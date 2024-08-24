// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"errors"
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type On any

type ActivityConfig struct {
	Id    string   `hcl:"id,label"`
	Types []string `hcl:"types,attr"`
}

type EventConfig struct {
	Event          string    `hcl:"event,label"`
	Branches       *[]string `hcl:"branches,attr"`
	BranchesIgnore *[]string `hcl:"branches_ignore,attr"`
	Tags           *[]string `hcl:"tags,attr"`
	TagsIgnore     *[]string `hcl:"tags_ignore,attr"`
	Paths          *[]string `hcl:"paths,attr"`
	PathsIgnore    *[]string `hcl:"paths_ignore,attr"`
}

type OnsConfig []OnConfig

type OnConfig struct {
	Events   *cty.Value        `hcl:"events,attr"`
	Event    []*EventConfig    `hcl:"event,block"`
	Activity []*ActivityConfig `hcl:"activity,block"`
}

func (config *OnConfig) parseActivity() (On, error) {
	activities := make(map[string]map[string][]string)

	for _, activity := range config.Activity {
		if !ValidateEventTrigger(activity.Id) {
			return nil, fmt.Errorf("invalid event trigger `%s`", activity.Id)
		}

		if activities[activity.Id] == nil {
			activities[activity.Id] = make(map[string][]string)
		}

		for _, activityType := range activity.Types {
			if !ValidateActivityType(activityType) {
				return nil, fmt.Errorf("invalid activity type `%s`", activityType)
			}
		}

		activities[activity.Id]["types"] = activity.Types
	}

	return activities, nil
}

func (config *OnConfig) parseEvent() (On, error) {
	events := make(map[string]map[string][]string)

	for _, event := range config.Event {
		if !ValidateEventTrigger(event.Event) {
			return nil, fmt.Errorf("invalid event trigger `%s`", event.Event)
		}

		if events[event.Event] == nil {
			events[event.Event] = make(map[string][]string)
		}

		if event.Branches != nil {
			events[event.Event]["branches"] = *event.Branches
		} else if event.BranchesIgnore != nil {
			events[event.Event]["branches-ignore"] = *event.BranchesIgnore
		}

		if event.Tags != nil {
			events[event.Event]["tags"] = *event.Tags
		} else if event.TagsIgnore != nil {
			events[event.Event]["tags-ignore"] = *event.TagsIgnore
		}

		if event.Paths != nil {
			events[event.Event]["paths"] = *event.Paths
		} else if event.PathsIgnore != nil {
			events[event.Event]["paths-ignore"] = *event.PathsIgnore
		}
	}

	return events, nil
}

func (config *OnConfig) parseEvents() (On, error) {
	events, err := ParseCtyValue(*config.Events, []string{
		cty.String.FriendlyName(),
		cty.EmptyTuple.FriendlyName(),
	})
	if err != nil {
		return nil, err
	}

	switch event := events.(type) {
	case string:
		if !ValidateEventTrigger(event) {
			return nil, fmt.Errorf("invalid event trigger `%s`", event)
		}
	case []string:
		for _, evt := range event {
			if !ValidateEventTrigger(evt) {
				return nil, fmt.Errorf("invalid event trigger `%s`", evt)
			}
		}
	}

	return events, nil
}

func (config *OnConfig) Parse() (On, error) {
	if config == nil {
		return nil, nil
	}

	if config.Events != nil && config.Event != nil {
		return nil, errors.New("on can only have Events or Event")
	}

	ons := make(map[string]map[string][]string)

	if config.Events != nil {
		on, err := config.parseEvents()
		if err != nil {
			return nil, err
		}
		return on, nil
	} else {
		partialOn, err := config.parseEvent()
		if err != nil {
			return nil, err
		}

		switch onValue := partialOn.(type) {
		case map[string]map[string][]string:
			for k, v := range onValue {
				ons[k] = v
			}
		default:
			return nil, errors.New("undefined error")
		}

	}

	if config.Activity != nil {
		partialOn, err := config.parseActivity()
		if err != nil {
			return nil, err
		}

		switch onValue := partialOn.(type) {
		case map[string]map[string][]string:
			for k, v := range onValue {
				ons[k] = v
			}
		}
	}

	return ons, nil
}

func (config *OnsConfig) Parse() (On, error) {
	ons := make(map[string]map[string][]string)
	for _, on := range *config {
		parsedOn, err := on.Parse()
		if err != nil {
			return nil, err
		}

		switch val := parsedOn.(type) {
		case string:
			return parsedOn, nil
		case []string:
			return parsedOn, nil
		case map[string]map[string][]string:
			for k, v := range val {
				ons[k] = v
			}
		default:
			return On(""), errors.New("unknown `on` structure")
		}
	}

	return ons, nil
}

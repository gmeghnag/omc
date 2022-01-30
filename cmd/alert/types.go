/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package alert

import "time"

type status string

var GroupFilename, RuleState, GroupName string

const (
	statusSuccess status = "success"
	statusError   status = "error"
)

type alerts struct {
	Status status `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	Groups []RuleGroup `json:"groups"`
}

type RuleGroup struct {
	Name string `json:"name"`
	File string `json:"file"`
	// In order to preserve rule ordering, while exposing type (alerting or recording)
	// specific properties, both alerting and recording rules are exposed in the
	// same array.
	Rules          []Rule    `json:"rules"`
	Interval       float64   `json:"interval"`
	Limit          int       `json:"limit"`
	EvaluationTime float64   `json:"evaluationTime"`
	LastEvaluation time.Time `json:"lastEvaluation"`
}

type Rule map[string]interface{}

type AlertingRule struct {
	// State can be "pending", "firing", "inactive".
	State          string       `json:"state"`
	Name           string       `json:"name"`
	Query          string       `json:"query"`
	Duration       float64      `json:"duration"`
	Labels         Labels       `json:"labels"`
	Annotations    Labels       `json:"annotations"`
	Alerts         []*PromAlert `json:"alerts"`
	Health         RuleHealth   `json:"health"`
	LastError      string       `json:"lastError,omitempty"`
	EvaluationTime float64      `json:"evaluationTime"`
	LastEvaluation time.Time    `json:"lastEvaluation"`
	// Type of an alertingRule is always "alerting".
	Type string `json:"type"`
}

type Label struct {
	Name, Value string
}

// Labels is a sorted set of labels. Order has to be guaranteed upon
// instantiation.
type Labels []Label

type PromAlert struct {
	Labels      Labels     `json:"labels"`
	Annotations Labels     `json:"annotations"`
	State       string     `json:"state"`
	ActiveAt    *time.Time `json:"activeAt,omitempty"`
	Value       string     `json:"value"`
}

type RuleHealth string

// The possible health states of a rule based on the last execution.
const (
	HealthUnknown RuleHealth = "unknown"
	HealthGood    RuleHealth = "ok"
	HealthBad     RuleHealth = "err"
)

type FilteredRulesList struct {
	Data []Rule `json:"data"`
}

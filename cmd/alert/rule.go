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

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func GetAlertRules(resourcesNames []string, outputFlag string, groupsNames string, rulesStates string, alertsFilePath string) {
	_headers := []string{"group", "rule", "state", "age", "alerts", "active since"}
	var data [][]string
	var filteredRules []Rule
	var filteredRulesList FilteredRulesList
	var _Alerts alerts
	_file, _ := ioutil.ReadFile(alertsFilePath)
	if err := yaml.Unmarshal([]byte(_file), &_Alerts); err != nil {
		fmt.Println("Error when trying to unmarshal file " + alertsFilePath)
		os.Exit(1)
	}
	searchingGroups := []string{}
	if groupsNames != "" {
		searchingGroups = strings.Split(groupsNames, ",")
	}
	searchingStates := []string{}
	if rulesStates != "" {
		searchingStates = strings.Split(rulesStates, ",")
	}

	for _, group := range _Alerts.Data.Groups {
		if len(searchingGroups) != 0 && !helpers.StringInSlice(group.Name, searchingGroups) {
			continue
		}

		for _, rule := range group.Rules {
			ruleName := fmt.Sprint(rule["name"])
			if len(resourcesNames) != 0 && !helpers.StringInSlice(ruleName, resourcesNames) {
				continue
			}
			ruleState := fmt.Sprint(rule["state"])
			if len(searchingStates) != 0 && !helpers.StringInSlice(ruleState, searchingStates) {
				continue
			}

			if outputFlag == "yaml" || outputFlag == "json" {
				filteredRules = append(filteredRules, rule)
				continue
			}

			activeSince := "----"
			// I didn't found any other solution than this (Marshal and Unmarshal) to transform alerts interface{} to []PromAlert{} :/
			alerts := rule["alerts"]
			alertsList := []PromAlert{}
			b, err := json.Marshal(alerts)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			json.Unmarshal(b, &alertsList)
			numAlerts := strconv.Itoa(len(alertsList))
			if len(alertsList) != 0 {
				if len(alertsList) > 1 {
					firstOccur := alertsList[0].ActiveAt
					for i := range alertsList[1:] {
						alertBefore := alertsList[i].ActiveAt
						if alertsList[i+1].ActiveAt.Before(*alertBefore) {
							firstOccur = alertsList[i+1].ActiveAt
						}
					}
					activeSince = firstOccur.Format(time.RFC822)
				} else {
					alert := alertsList[0]
					activeSince = alert.ActiveAt.Format(time.RFC822)
				}
			}
			ruleLastEvaluation := fmt.Sprint(rule["lastEvaluation"])
			ruleLastEvaluationTime, _ := time.Parse(time.RFC3339Nano, ruleLastEvaluation)
			ResourceFile, _ := os.Stat(alertsFilePath)
			t2 := ResourceFile.ModTime()
			diffTime := t2.Sub(ruleLastEvaluationTime).String()
			d, _ := time.ParseDuration(diffTime)
			lastEval := helpers.FormatDiffTime(d)
			_list := []string{group.Name, ruleName, ruleState, lastEval, numAlerts, activeSince}
			showGroup := false
			if outputFlag == "wide" {
				showGroup = true
			}
			data = helpers.GetData(data, showGroup, false, "", outputFlag, 6, _list)
		}
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[1:]
		if len(data) == 0 {
			fmt.Println("No resources found.")
		} else {
			helpers.PrintTable(headers, data)
		}
	}
	if outputFlag == "wide" {
		headers = _headers[0:]
		if len(data) == 0 {
			fmt.Println("No resources found.")
		} else {
			helpers.PrintTable(headers, data)
		}
	}
	if outputFlag == "yaml" {
		filteredRulesList.Data = filteredRules
		y, _ := yaml.Marshal(filteredRulesList)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		filteredRulesList.Data = filteredRules
		j, _ := json.Marshal(filteredRulesList)
		fmt.Println(string(j))
	}
}

var RuleSubCmd = &cobra.Command{
	Use:     "rule",
	Aliases: []string{"rules"},
	Run: func(cmd *cobra.Command, args []string) {
		resourcesNames := args
		monitoringExist, _ := helpers.Exists(vars.MustGatherRootPath + "/monitoring")
		if !monitoringExist {
			fmt.Println("Path '" + vars.MustGatherRootPath + "/monitoring' does not exist.")
			os.Exit(1)
		}
		alertsFilePath := vars.MustGatherRootPath + "/monitoring/alerts.json"
		alertsFilePathExist, _ := helpers.Exists(alertsFilePath)
		if !alertsFilePathExist {
			alertsFilePath = vars.MustGatherRootPath + "/monitoring/prometheus/rules.json"
			alertsFilePathExist, _ := helpers.Exists(alertsFilePath)
			if !alertsFilePathExist {
				fmt.Println("Prometheus rules not found in must-gather.")
				os.Exit(1)
			}
		}
		GetAlertRules(resourcesNames, vars.OutputStringVar, GroupName, RuleState, alertsFilePath)
	},
}

func init() {
	RuleSubCmd.Flags().StringVarP(&GroupName, "group", "g", "", "Filter the AlertRules by AlertGroup/s (comma separated).")
	RuleSubCmd.Flags().StringVarP(&RuleState, "state", "s", "", "Filter the AlertRules by state.")
}

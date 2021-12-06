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
package batch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"
)

type CronJobsItems struct {
	ApiVersion string             `json:"apiVersion"`
	Items      []*batchv1.CronJob `json:"items"`
}

func GetCronJobs(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "schedule", "suspend", "active", "last schedule", "age", "containers", "images", "selector"}

	var namespaces []string
	if allNamespacesFlag == true {
		namespace = "all"
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, namespace)
	}

	var data [][]string
	var _CronJobsList = CronJobsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items CronJobsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/batch/cronjobs.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/batch/cronjobs.yaml")
			os.Exit(1)
		}

		for _, CronJob := range _Items.Items {
			if resourceName != "" && resourceName != CronJob.Name {
				continue
			}

			if outputFlag == "yaml" {
				_CronJobsList.Items = append(_CronJobsList.Items, CronJob)
				continue
			}

			if outputFlag == "json" {
				_CronJobsList.Items = append(_CronJobsList.Items, CronJob)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_CronJobsList.Items = append(_CronJobsList.Items, CronJob)
				continue
			}

			//name
			CronJobName := CronJob.Name
			if allResources {
				CronJobName = "cronjob.batch/" + CronJobName
			}
			//schedule
			schedule := CronJob.Spec.Schedule
			//suspend
			suspend := ""
			if CronJob.Spec.Suspend != nil {
				suspend = strconv.FormatBool(*CronJob.Spec.Suspend)
			}
			//active
			active := strconv.Itoa(len(CronJob.Status.Active))
			// last schedule
			lastSchedule := "<none>"
			if CronJob.Status.LastScheduleTime != nil {
				lastSchedule = helpers.GetAge(CurrentNamespacePath+"/batch/cronjobs.yaml", *CronJob.Status.LastScheduleTime)
			}

			//age
			age := helpers.GetAge(CurrentNamespacePath+"/batch/cronjobs.yaml", CronJob.GetCreationTimestamp())

			containers := ""
			for _, c := range CronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
				containers += fmt.Sprint(c.Name) + ","
			}
			if containers == "" {
				containers = "??"
			} else {
				containers = strings.TrimRight(containers, ",")
			}
			//images
			images := ""
			for _, i := range CronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
				images += fmt.Sprint(i.Image) + ","
			}
			if images == "" {
				images = "??"
			} else {
				images = strings.TrimRight(images, ",")
			}
			selector := "<none>"
			if CronJob.Spec.JobTemplate.Spec.Selector != nil {
				for k, v := range CronJob.Spec.JobTemplate.Spec.Selector.MatchLabels {
					selector += k + "=" + v + ","
				}
				if selector == "" {
					selector = "<none>"
				} else {
					selector = strings.TrimRight(selector, ",")
				}
			}
			//labels
			labels := helpers.ExtractLabels(CronJob.GetLabels())
			_list := []string{CronJob.Namespace, CronJobName, schedule, suspend, active, lastSchedule, age, containers, images, selector}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 7, _list)

			if resourceName != "" && resourceName == CronJobName {
				break
			}
		}
		if namespace != "" && _namespace == namespace {
			break
		}
	}

	if (outputFlag == "" || outputFlag == "wide") && len(data) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var headers []string
	if outputFlag == "" {
		if allNamespacesFlag == true {
			headers = _headers[0:7]
		} else {
			headers = _headers[1:7]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		if allNamespacesFlag == true {
			headers = _headers
		} else {
			headers = _headers[1:]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}

	if len(_CronJobsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _CronJobsList.Items[0]
	} else {
		resource = _CronJobsList
	}
	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(resource)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(resource, "", "  ")
		fmt.Println(string(j))

	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(resource, jsonPathTemplate)
	}
	return false
}

var CronJob = &cobra.Command{
	Use:     "cronjob",
	Aliases: []string{"cronjobs", "cronjob.batch"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetCronJobs(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

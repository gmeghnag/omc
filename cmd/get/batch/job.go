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
	"os"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"
)

type JobsItems struct {
	ApiVersion string         `json:"apiVersion"`
	Items      []*batchv1.Job `json:"items"`
}

func GetJobs(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "completions", "duration", "age"}

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
	var _JobsList = JobsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items JobsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/batch/jobs.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/batch/jobs.yaml")
			os.Exit(1)
		}

		for _, Job := range _Items.Items {
			if resourceName != "" && resourceName != Job.Name {
				continue
			}

			if outputFlag == "yaml" {
				_JobsList.Items = append(_JobsList.Items, Job)
				continue
			}

			if outputFlag == "json" {
				_JobsList.Items = append(_JobsList.Items, Job)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_JobsList.Items = append(_JobsList.Items, Job)
				continue
			}

			//name
			JobName := Job.Name
			if allResources {
				JobName = "job.batch/" + JobName
			}
			//completions
			//fmt.Println(strconv.Itoa(int(*Job.Spec.Completions)))
			completions := "" //strconv.Itoa(int(Job.Status.Succeeded)) + "/" + strconv.Itoa(int(*Job.Spec.Completions))
			if Job.Spec.Completions != nil {
				completions = "" //strconv.Itoa(int(Job.Status.Succeeded)) + "/" + strconv.Itoa(int(*Job.Spec.Completions))
			}
			//duration
			duration := "Unknown"
			if Job.Status.CompletionTime != nil {
				t2 := Job.Status.CompletionTime.Time
				diffTime := t2.Sub(Job.Status.StartTime.Time).String()
				d, _ := time.ParseDuration(diffTime)
				duration = helpers.FormatDiffTime(d)
			}

			//age
			age := helpers.GetAge(CurrentNamespacePath+"/batch/jobs.yaml", Job.GetCreationTimestamp())

			//labels
			labels := helpers.ExtractLabels(Job.GetLabels())
			_list := []string{Job.Namespace, JobName, completions, duration, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == JobName {
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
			headers = _headers[0:5]
		} else {
			headers = _headers[1:5]
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

	if len(_JobsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _JobsList.Items[0]
	} else {
		resource = _JobsList
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

var Job = &cobra.Command{
	Use:     "job",
	Aliases: []string{"jobs", "job.batch"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetJobs(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

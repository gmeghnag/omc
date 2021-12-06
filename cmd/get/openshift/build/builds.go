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
package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	v1 "github.com/openshift/api/build/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type BuildsItems struct {
	ApiVersion string      `json:"apiVersion"`
	Items      []*v1.Build `json:"items"`
}

func GetBuilds(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "type", "from", "status", "started", "duration"}

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
	var _BuildsList = BuildsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items BuildsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/build.openshift.io/builds.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/build.openshift.io/builds.yaml")
			os.Exit(1)
		}

		for _, Build := range _Items.Items {
			if resourceName != "" && resourceName != Build.Name {
				continue
			}

			if outputFlag == "yaml" {
				_BuildsList.Items = append(_BuildsList.Items, Build)
				continue
			}

			if outputFlag == "json" {
				_BuildsList.Items = append(_BuildsList.Items, Build)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_BuildsList.Items = append(_BuildsList.Items, Build)
				continue
			}

			//name
			BuildName := Build.Name
			if allResources {
				BuildName = "build.build.openshift.io/" + BuildName
			}
			//type
			bcType := string(Build.Spec.Strategy.Type)
			//from
			from := string(Build.Spec.Source.Type)
			if Build.Spec.Revision != nil {
				if Build.Spec.Revision.Type == "Git" {
					from += "@" + Build.Spec.Revision.Git.Commit[0:7]
				}
			}
			//status
			status := string(Build.Status.Phase)
			//started
			started := helpers.GetAge(CurrentNamespacePath+"/build.openshift.io/builds.yaml", *Build.Status.StartTimestamp)
			//duration
			duration := strconv.Itoa(int(Build.Status.Duration/1000000000)) + "s"
			//labels
			labels := helpers.ExtractLabels(Build.GetLabels())
			_list := []string{Build.Namespace, BuildName, bcType, from, status, started, duration}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 7, _list)

			if resourceName != "" && resourceName == BuildName {
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

	if len(_BuildsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = _BuildsList.Items[0]
	} else {
		resource = _BuildsList
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

var Build = &cobra.Command{
	Use:     "build",
	Aliases: []string{"builds"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetBuilds(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

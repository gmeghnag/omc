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
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strconv"
	"strings"

	v1 "github.com/openshift/api/build/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type BuildConfigsItems struct {
	ApiVersion string            `json:"apiVersion"`
	Items      []*v1.BuildConfig `json:"items"`
}

func getBuildConfigs(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "type", "from", "latest"}

	var namespaces []string
	if allNamespacesFlag == true {
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	}
	if namespace != "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}
	if namespace == "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}

	var data [][]string
	var _BuildConfigsList = BuildConfigsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items BuildConfigsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/build.openshift.io/buildconfigs.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/build.openshift.io/buildconfigs.yaml")
			os.Exit(1)
		}

		for _, BuildConfig := range _Items.Items {
			if resourceName != "" && resourceName != BuildConfig.Name {
				continue
			}

			if outputFlag == "yaml" {
				_BuildConfigsList.Items = append(_BuildConfigsList.Items, BuildConfig)
				continue
			}

			if outputFlag == "json" {
				_BuildConfigsList.Items = append(_BuildConfigsList.Items, BuildConfig)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_BuildConfigsList.Items = append(_BuildConfigsList.Items, BuildConfig)
				continue
			}

			//name
			BuildConfigName := BuildConfig.Name
			if allResources {
				BuildConfigName = "buildconfig.build.openshift.io/" + BuildConfigName
			}
			//type
			bcType := string(BuildConfig.Spec.Strategy.Type)
			//from
			from := string(BuildConfig.Spec.Source.Type)
			//latest
			latest := strconv.Itoa(int(BuildConfig.Status.LastVersion))
			//labels
			labels := helpers.ExtractLabels(BuildConfig.GetLabels())
			_list := []string{BuildConfig.Namespace, BuildConfigName, bcType, from, latest}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == BuildConfigName {
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

	if len(_BuildConfigsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = _BuildConfigsList.Items[0]
	} else {
		resource = _BuildConfigsList
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

var BuildConfig = &cobra.Command{
	Use:     "buildconfig",
	Aliases: []string{"buildconfigs", "bc"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getBuildConfigs(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

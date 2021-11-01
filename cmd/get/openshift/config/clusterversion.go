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
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"reflect"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"
)

type ClusterVersionsItems struct {
	ApiVersion string                    `json:"apiVersion"`
	Items      []configv1.ClusterVersion `json:"items"`
}

func getClusterVersion(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	clusterversionsFolderPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/clusterversions/"
	_clusterversions, _ := ioutil.ReadDir(clusterversionsFolderPath)

	_headers := []string{"name", "version", "available", "progressing", "since", "status"}
	var data [][]string

	_ClusterVersionsList := ClusterVersionsItems{ApiVersion: "v1"}
	for _, f := range _clusterversions {
		clusterversionYamlPath := clusterversionsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(clusterversionYamlPath)
		ClusterVersion := configv1.ClusterVersion{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterVersion); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + clusterversionYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != ClusterVersion.Name {
			continue
		}

		if outputFlag == "yaml" {
			_ClusterVersionsList.Items = append(_ClusterVersionsList.Items, ClusterVersion)
			continue
		}

		if outputFlag == "json" {
			_ClusterVersionsList.Items = append(_ClusterVersionsList.Items, ClusterVersion)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ClusterVersionsList.Items = append(_ClusterVersionsList.Items, ClusterVersion)
			continue
		}
		//Name
		clusterOperatorName := ClusterVersion.Name
		//version
		version := ""
		for _, h := range ClusterVersion.Status.History {
			if h.State == "Completed" {
				version = h.Version
			}
		}
		// conditions
		conditions := ClusterVersion.Status.Conditions
		available := ""
		progressing := ""
		status := ""
		var lastsTransitionTime []v1.Time
		var lastTransitionTime v1.Time
		var zeroTime v1.Time
		for _, c := range conditions {
			//available
			if c.Type == "Available" {
				available = string(c.Status)
				lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
			}
			//progressing
			if c.Type == "Progressing" {
				progressing = string(c.Status)
				status = string(c.Message)
				lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
			}
			//status
			if c.Type == "Failing" {
				lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
			}
		}
		//since
		for _, t := range lastsTransitionTime {
			if reflect.DeepEqual(lastTransitionTime, zeroTime) {
				lastTransitionTime = t
			} else {
				if t.Time.After(lastTransitionTime.Time) {
					lastTransitionTime = t
				}
			}
		}
		since := helpers.GetAge(clusterversionYamlPath, lastTransitionTime)
		labels := helpers.ExtractLabels(ClusterVersion.GetLabels())
		_list := []string{clusterOperatorName, version, available, progressing, since, status}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 6, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:6] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)

	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
	}
	var resource interface{}
	if resourceName != "" {
		resource = _ClusterVersionsList.Items[0]
	} else {
		resource = _ClusterVersionsList
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

var ClusterVersion = &cobra.Command{
	Use:     "clusterversion",
	Aliases: []string{"clusterversions"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterVersion(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

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
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"
)

type ClusterOperatorsItems struct {
	ApiVersion string                     `json:"apiVersion"`
	Items      []configv1.ClusterOperator `json:"items"`
}

func getClusterOperators(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	clusteroperatorsFolderPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/clusteroperators/"
	_clusteroperators, _ := ioutil.ReadDir(clusteroperatorsFolderPath)

	_headers := []string{"name", "version", "available", "progressing", "degraded", "since"}
	var data [][]string

	_ClusterOperatorsList := ClusterOperatorsItems{ApiVersion: "v1"}
	for _, f := range _clusteroperators {
		clusteroperatorYamlPath := clusteroperatorsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(clusteroperatorYamlPath)
		ClusterOperator := configv1.ClusterOperator{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterOperator); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + clusteroperatorYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(ClusterOperator.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != ClusterOperator.Name {
			continue
		}

		if outputFlag == "name" {
			_ClusterOperatorsList.Items = append(_ClusterOperatorsList.Items, ClusterOperator)
			fmt.Println("clusteroperator.config.openshift.io/" + ClusterOperator.Name)
			continue
		}

		if outputFlag == "yaml" {
			_ClusterOperatorsList.Items = append(_ClusterOperatorsList.Items, ClusterOperator)
			continue
		}

		if outputFlag == "json" {
			_ClusterOperatorsList.Items = append(_ClusterOperatorsList.Items, ClusterOperator)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ClusterOperatorsList.Items = append(_ClusterOperatorsList.Items, ClusterOperator)
			continue
		}
		//Name
		clusterOperatorName := ClusterOperator.Name
		//version
		version := ""
		for _, v := range ClusterOperator.Status.Versions {
			if v.Name == "operator" {
				version = v.Version
			}
		}
		// conditions
		conditions := ClusterOperator.Status.Conditions
		available := ""
		progressing := ""
		degraded := ""
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
				lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
			}
			//degraded
			if c.Type == "Degraded" {
				degraded = string(c.Status)
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
		since := "??"
		ResourceFile, _ := os.Stat(clusteroperatorYamlPath)
		t2 := ResourceFile.ModTime()
		diffTime := t2.Sub(lastTransitionTime.Time).String()
		d, _ := time.ParseDuration(diffTime)
		since = helpers.FormatDiffTime(d)
		_list := []string{clusterOperatorName, version, available, progressing, degraded, since}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 6, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:6] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false

	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	var resource interface{}
	if resourceName != "" {
		resource = _ClusterOperatorsList.Items[0]
	} else {
		resource = _ClusterOperatorsList
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

var ClusterOperator = &cobra.Command{
	Use:     "clusteroperator",
	Aliases: []string{"co", "clusteroperators", "clusteroperator.config.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterOperators(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

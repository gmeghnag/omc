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

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"
)

type ClusterVersionsItems struct {
	ApiVersion string                    `json:"apiVersion"`
	Items      []configv1.ClusterVersion `json:"items"`
}

func getClusterVersionV2(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

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
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+clusterversionYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(ClusterVersion.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != ClusterVersion.Name {
			continue
		}

		if outputFlag == "name" {
			_ClusterVersionsList.Items = append(_ClusterVersionsList.Items, ClusterVersion)
			fmt.Println("clusterversion.config.openshift.io/" + ClusterVersion.Name)
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
				break
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

func getClusterVersionV1(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	clusterversionsFileExist, _ := helpers.Exists(currentContextPath + "/cluster-scoped-resources/config.openshift.io/clusterversions.yaml")
	clusterversionsYamlPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/clusterversions.yaml"
	clusterversions := ClusterVersionsItems{ApiVersion: "v1"}
	if clusterversionsFileExist {
		_file, _ := ioutil.ReadFile(clusterversionsYamlPath)
		if err := yaml.Unmarshal([]byte(_file), &clusterversions); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+clusterversionsYamlPath)
			os.Exit(1)
		}
	}

	_headers := []string{"name", "version", "available", "progressing", "since", "status"}
	var data [][]string

	_ClusterVersionsList := ClusterVersionsItems{ApiVersion: "v1"}
	for _, ClusterVersion := range clusterversions.Items {

		labels := helpers.ExtractLabels(ClusterVersion.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != ClusterVersion.Name {
			continue
		}

		if outputFlag == "name" {
			_ClusterVersionsList.Items = append(_ClusterVersionsList.Items, ClusterVersion)
			fmt.Println("clusterversion.config.openshift.io/" + ClusterVersion.Name)
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
				break
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
		since := helpers.GetAge(clusterversionsYamlPath, lastTransitionTime)

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
	Aliases: []string{"clusterversions", "clusterversion.config.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		clusterversionsFileExist, _ := helpers.Exists(vars.MustGatherRootPath + "/cluster-scoped-resources/config.openshift.io/clusterversions.yaml")
		if clusterversionsFileExist {
			getClusterVersionV1(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
		} else {
			getClusterVersionV2(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
		}
	},
}

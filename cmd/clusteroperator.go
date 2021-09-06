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
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"omc/cmd/helpers"
	"os"
	"strings"

	configv1 "github.com/openshift/api/config/v1"

	"sigs.k8s.io/yaml"
)

type ClusterOperatorsItems struct {
	ApiVersion string                     `json:"apiVersion"`
	Items      []configv1.ClusterOperator `json:"items"`
}

func getClusterOperators(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) {
	// get quay-io-... string
	files, err := ioutil.ReadDir(currentContextPath)
	if err != nil {
		log.Fatal(err)
	}
	var QuayString string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "quay") {
			QuayString = f.Name()
			break
		}
	}
	if QuayString == "" {
		fmt.Println("Some error occurred, wrong must-gather file composition")
		os.Exit(1)
	}

	clusteroperatorsFolderPath := currentContextPath + "/" + QuayString + "/cluster-scoped-resources/config.openshift.io/clusteroperators/"
	_clusteroperators, _ := ioutil.ReadDir(clusteroperatorsFolderPath)

	_headers := []string{"name", "version", "available", "progressing", "degraded", "since"}
	var data [][]string

	_ClusterOperatorsList := ClusterOperatorsItems{ApiVersion: "v1"}
	for _, f := range _clusteroperators {
		clusteroperatorYamlPath := clusteroperatorsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(clusteroperatorYamlPath)
		ClusterOperator := configv1.ClusterOperator{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterOperator); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + clusteroperatorYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != ClusterOperator.Name {
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
		version := ClusterOperator.Status.Versions[0].Version // TODO why the list?
		//available
		available := "??"
		//progressing
		progressing := "??"
		//degraded
		degraded := "??"
		//since
		since := "??"
		labels := helpers.ExtractLabels(ClusterOperator.GetLabels())
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

}

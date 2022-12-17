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
package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func getClusterLoggings(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	ClusterloggingYamlPath := currentContextPath + "/cluster-logging/clo/cr"
	_headers := []string{"name", "age"}
	var data [][]string

	_file := helpers.ReadYaml(ClusterloggingYamlPath)
	ClusterLogging := logging.ClusterLogging{}
	if err := yaml.Unmarshal([]byte(_file), &ClusterLogging); err != nil {
		fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+ClusterloggingYamlPath)
		os.Exit(1)
	}

	labels := helpers.ExtractLabels(ClusterLogging.GetLabels())
	if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
		return false
	}
	if resourceName != "" && resourceName != ClusterLogging.Name {
		return false
	}

	if outputFlag == "name" {
		fmt.Println("Clusterlogging/" + ClusterLogging.Name)
		return false
	}

	//Name
	ClusterloggingName := ClusterLogging.Name
	age := helpers.GetAge(ClusterloggingYamlPath, ClusterLogging.GetCreationTimestamp())

	_list := []string{ClusterloggingName, age}
	data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:2] // -A
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

	// TODO: implement --show-managed-fields bool flags
	ClusterLogging.ObjectMeta.ManagedFields = nil

	if outputFlag == "yaml" {

		y, _ := yaml.Marshal(ClusterLogging)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(ClusterLogging, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(ClusterLogging, jsonPathTemplate)
	}
	return false
}

var ClusterLogging = &cobra.Command{
	Use:     "clusterlogging",
	Aliases: []string{"clusterloggings", "cl"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterLoggings(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

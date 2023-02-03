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

type ClusterLogForwarderItems struct {
	ApiVersion string                        `json:"apiVersion"`
	Items      []logging.ClusterLogForwarder `json:"items"`
	Kind       string                        `json:"kind"`
}

func getClusterLogForwarders(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	clusterlogforwarderYamlPath := currentContextPath + "/cluster-logging/clo/clusterlogforwarder_instance.yaml"
	_headers := []string{"name", "age"}
	var data [][]string

	_file := helpers.ReadYaml(clusterlogforwarderYamlPath)
	ClusterLogForwarderList := ClusterLogForwarderItems{ApiVersion: "v1", Kind: "List"}
	
	if err := yaml.Unmarshal([]byte(_file), &ClusterLogForwarderList); err != nil {
		fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+clusterlogforwarderYamlPath)
		os.Exit(1)
	}

	ClusterLogForwarder := ClusterLogForwarderList.Items[0]
	labels := helpers.ExtractLabels(ClusterLogForwarder.GetLabels())
	if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
		return false
	}
	if resourceName != "" && resourceName != ClusterLogForwarder.Name {
		return false
	}

	if outputFlag == "name" {
		fmt.Println("clusterlogforwarder/" + ClusterLogForwarder.Name)
		return false
	}

	//Name
	clusterlogforwarderName := ClusterLogForwarder.Name
	age := helpers.GetAge(clusterlogforwarderYamlPath, ClusterLogForwarder.GetCreationTimestamp())

	_list := []string{clusterlogforwarderName, age}
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
	ClusterLogForwarder.ObjectMeta.ManagedFields = nil

	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(ClusterLogForwarder)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(ClusterLogForwarder, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(ClusterLogForwarder, jsonPathTemplate)
	}
	return false
}

var ClusterLogForwarder = &cobra.Command{
	Use:     "clusterlogforwarder",
	Aliases: []string{"clusterlogforwarders", "clf"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterLogForwarders(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

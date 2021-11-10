/*
Copyright © 2021 Bram Verschueren <bverschueren@redhat.com>

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
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

type NetworksItems struct {
	ApiVersion string             `json:"apiVersion"`
	Items      []configv1.Network `json:"items"`
}

func getNetwork(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	networksYamlPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/networks.yaml"

	_headers := []string{"name", "age"}
	var data [][]string

	_file, _ := ioutil.ReadFile(networksYamlPath)
	NetworkList := configv1.NetworkList{}
	if err := yaml.Unmarshal([]byte(_file), &NetworkList); err != nil {
		fmt.Println("Error when trying to unmarshall file: " + networksYamlPath)
		os.Exit(1)
	}

	_NetworksList := NetworksItems{ApiVersion: "v1"}

	for _, Network := range NetworkList.Items {

		if resourceName != "" && resourceName != Network.Name {
			continue
		}

		_NetworksList.Items = append(_NetworksList.Items, Network)

		NetworkName := Network.Name
		since := helpers.GetAge(networksYamlPath, Network.GetCreationTimestamp())
		labels := helpers.ExtractLabels(Network.GetLabels())
		_list := []string{NetworkName, since}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:2] // -A
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
		resource = _NetworksList.Items[0]
	} else {
		resource = _NetworksList
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

var Network = &cobra.Command{
	Use:     "network",
	Aliases: []string{"networks", "network.config.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getNetwork(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

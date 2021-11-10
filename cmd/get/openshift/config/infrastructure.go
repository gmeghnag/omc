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

func getInfrastructures(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	infrastructuresFolderPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/infrastructures/"
	_infrastructures, _ := ioutil.ReadDir(infrastructuresFolderPath)

	_headers := []string{"name", "age"}
	var data [][]string

	_InfrastructuresList := configv1.InfrastructureList{}
	for _, f := range _infrastructures {
		infrastructureYamlPath := infrastructuresFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(infrastructureYamlPath)
		infrastructure := configv1.Infrastructure{}
		if err := yaml.Unmarshal([]byte(_file), &infrastructure); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + infrastructureYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != infrastructure.Name {
			continue
		}

		_InfrastructuresList.Items = append(_InfrastructuresList.Items, infrastructure)

		//Name
		infrastructureName := infrastructure.Name
		age := helpers.GetAge(infrastructureYamlPath, infrastructure.GetCreationTimestamp())

		labels := helpers.ExtractLabels(infrastructure.GetLabels())
		_list := []string{infrastructureName, age}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)
	}

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
	var resource interface{}
	if resourceName != "" {
		resource = _InfrastructuresList.Items[0]
	} else {
		resource = _InfrastructuresList
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

var Infrastructure = &cobra.Command{
	Use:     "infrastructure",
	Aliases: []string{"infrastructure", "infrastructure.config.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getInfrastructures(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

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
package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type ConfigMapsItems struct {
	ApiVersion string              `json:"apiVersion"`
	Items      []*corev1.ConfigMap `json:"items"`
}

func getConfigMaps(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "data", "age"}
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
	var _ConfigMapsList = ConfigMapsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items ConfigMapsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/configmaps.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/configmaps.yaml")
			os.Exit(1)
		}

		for _, ConfigMap := range _Items.Items {
			// configmap path
			if resourceName != "" && resourceName != ConfigMap.Name {
				continue
			}

			if outputFlag == "yaml" {
				_ConfigMapsList.Items = append(_ConfigMapsList.Items, ConfigMap)
				continue
			}

			if outputFlag == "json" {
				_ConfigMapsList.Items = append(_ConfigMapsList.Items, ConfigMap)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_ConfigMapsList.Items = append(_ConfigMapsList.Items, ConfigMap)
				continue
			}

			//name
			ConfigMapName := ConfigMap.Name
			if allResources {
				ConfigMapName = "configmap/" + ConfigMapName
			}
			//data
			configmapData := strconv.Itoa(len(ConfigMap.Data))

			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/configmaps.yaml", ConfigMap.GetCreationTimestamp())
			//labels
			labels := helpers.ExtractLabels(ConfigMap.GetLabels())
			_list := []string{ConfigMap.Namespace, ConfigMapName, configmapData, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 4, _list)

			if resourceName != "" && resourceName == ConfigMapName {
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
			headers = _headers[0:4]
		} else {
			headers = _headers[1:4]
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

	if len(_ConfigMapsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _ConfigMapsList.Items[0]
	} else {
		resource = _ConfigMapsList
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

	/* var resource interface{}
	if outputFlag == "yaml" || outputFlag == "json" || outputFlag == "jsonpath" {
		if resourceName != "" {
			fmt.Println(_ConfigMapsList.Items[0].Name)
			resource = _ConfigMapsList.Items[0]
		} else {
			resource = _ConfigMapsList
		}
	}
	helpers.PrintOutput(resource, outputFlag, resourceName, allNamespacesFlag, showLabels, _headers, data, jsonPathTemplate)
	*/return false
}

var ConfigMap = &cobra.Command{
	Use:     "configmap",
	Aliases: []string{"cm"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getConfigMaps(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

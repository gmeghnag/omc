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
package machineconfiguration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

type MachineConfigsItems struct {
	ApiVersion string                   `json:"apiVersion"`
	Items      []configv1.MachineConfig `json:"items"`
}

func getMachineConfig(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	machineconfigsFolderPath := currentContextPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigs/"
	_machineconfigs, _ := ioutil.ReadDir(machineconfigsFolderPath)
	_headers := []string{"name", "generatedbycontroller", "ignitionversion", "age"}
	var data [][]string

	_MachineConfigsList := MachineConfigsItems{ApiVersion: "v1"}
	for _, f := range _machineconfigs {
		machineconfigYamlPath := machineconfigsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(machineconfigYamlPath)
		MachineConfig := configv1.MachineConfig{}
		if err := yaml.Unmarshal([]byte(_file), &MachineConfig); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + machineconfigYamlPath)
			os.Exit(1)
		}
		if resourceName != "" && resourceName != MachineConfig.Name {
			continue
		}

		if outputFlag == "yaml" {
			_MachineConfigsList.Items = append(_MachineConfigsList.Items, MachineConfig)
			continue
		}

		if outputFlag == "json" {
			_MachineConfigsList.Items = append(_MachineConfigsList.Items, MachineConfig)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_MachineConfigsList.Items = append(_MachineConfigsList.Items, MachineConfig)
			continue
		}
		//Name
		MachineConfigName := MachineConfig.Name
		//generatedbycontroller
		generatedbycontroller := ""
		for i, k := range MachineConfig.ObjectMeta.Annotations {
			if strings.HasPrefix(i, "machineconfiguration.openshift.io/generated-by-controller-version") {
				generatedbycontroller = k
				break
			}
		}
		//ignitionversion
		jsonString := string(MachineConfig.Spec.Config.Raw[:])
		var jsonMap map[string]interface{}
		json.Unmarshal([]byte(jsonString), &jsonMap)
		ign := jsonMap["ignition"]
		ignMap := ign.(map[string]interface{})
		ignitionversion := fmt.Sprint(ignMap["version"])
		//age
		age := helpers.GetAge(machineconfigYamlPath, MachineConfig.GetCreationTimestamp())
		labels := helpers.ExtractLabels(MachineConfig.GetLabels())
		_list := []string{MachineConfigName, generatedbycontroller, ignitionversion, age}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 4, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:4] // -A
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
		resource = _MachineConfigsList.Items[0]
	} else {
		resource = _MachineConfigsList
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

var MachineConfig = &cobra.Command{
	Use:     "machineconfig",
	Aliases: []string{"machineconfigs", "mc"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getMachineConfig(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

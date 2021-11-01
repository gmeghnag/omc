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
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strconv"
	"strings"

	configv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

type MachineConfigPoolsItems struct {
	ApiVersion string                       `json:"apiVersion"`
	Items      []configv1.MachineConfigPool `json:"items"`
}

func getMachineConfigPool(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	machineconfigpoolsFolderPath := currentContextPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigpools/"
	_machineconfigpools, _ := ioutil.ReadDir(machineconfigpoolsFolderPath)

	_headers := []string{"name", "config", "updated", "updating", "degraded", "machinecount", "readymachinecount", "updatedmachinecount", "degradedmachinecount", "age"}
	var data [][]string

	_MachineConfigPoolsList := MachineConfigPoolsItems{ApiVersion: "v1"}
	for _, f := range _machineconfigpools {
		machineconfigpoolYamlPath := machineconfigpoolsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(machineconfigpoolYamlPath)
		MachineConfigPool := configv1.MachineConfigPool{}
		if err := yaml.Unmarshal([]byte(_file), &MachineConfigPool); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + machineconfigpoolYamlPath)
			os.Exit(1)
		}
		if resourceName != "" && resourceName != MachineConfigPool.Name {
			continue
		}

		if outputFlag == "yaml" {
			_MachineConfigPoolsList.Items = append(_MachineConfigPoolsList.Items, MachineConfigPool)
			continue
		}

		if outputFlag == "json" {
			_MachineConfigPoolsList.Items = append(_MachineConfigPoolsList.Items, MachineConfigPool)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_MachineConfigPoolsList.Items = append(_MachineConfigPoolsList.Items, MachineConfigPool)
			continue
		}
		//Name
		clusterOperatorName := MachineConfigPool.Name
		//config
		config := MachineConfigPool.Spec.Configuration.Name

		// conditions
		conditions := MachineConfigPool.Status.Conditions
		updated := ""
		updating := ""
		degraded := ""

		for _, c := range conditions {
			//updated
			if c.Type == "Updated" {
				updated = string(c.Status)
			}
			//updating
			if c.Type == "Updating" {
				updating = string(c.Status)
			}
			//degraded
			if c.Type == "Degraded" {
				degraded = string(c.Status)
			}
		}

		//machinecount
		machinecount := strconv.Itoa(int(MachineConfigPool.Status.MachineCount))
		//readymachinecount
		readymachinecount := strconv.Itoa(int(MachineConfigPool.Status.ReadyMachineCount))
		//updatedmachinecount
		updatedmachinecount := strconv.Itoa(int(MachineConfigPool.Status.UpdatedMachineCount))
		//degradedmachinecount
		degradedmachinecount := strconv.Itoa(int(MachineConfigPool.Status.DegradedMachineCount))
		//age
		age := helpers.GetAge(machineconfigpoolYamlPath, MachineConfigPool.GetCreationTimestamp())
		labels := helpers.ExtractLabels(MachineConfigPool.GetLabels())
		_list := []string{clusterOperatorName, config, updated, updating, degraded, machinecount, readymachinecount, updatedmachinecount, degradedmachinecount, age}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 10, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:10] // -A
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
		resource = _MachineConfigPoolsList.Items[0]
	} else {
		resource = _MachineConfigPoolsList
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

var MachineConfigPool = &cobra.Command{
	Use:     "machineconfigpool",
	Aliases: []string{"machineconfigpools", "mcp"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getMachineConfigPool(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

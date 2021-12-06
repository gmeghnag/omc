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
package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	machineapi "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type MachineSetsItems struct {
	ApiVersion string                  `json:"apiVersion"`
	Items      []machineapi.MachineSet `json:"items"`
}

func getMachineSets(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "desired", "current", "ready", "available", "age"}
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
	var _MachineSetsList = MachineSetsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_machinesets, err := ioutil.ReadDir(CurrentNamespacePath + "/machine.openshift.io/machinesets/")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		for _, f := range _machinesets {
			machineYamlPath := CurrentNamespacePath + "/machine.openshift.io/machinesets/" + f.Name()
			_file := helpers.ReadYaml(machineYamlPath)
			MachineSet := machineapi.MachineSet{}
			if err := yaml.Unmarshal([]byte(_file), &MachineSet); err != nil {
				fmt.Println("Error when trying to unmarshall file " + machineYamlPath)
				os.Exit(1)
			}

			// secret path
			if resourceName != "" && resourceName != MachineSet.Name {
				continue
			}

			if outputFlag == "yaml" {
				_MachineSetsList.Items = append(_MachineSetsList.Items, MachineSet)
				continue
			}

			if outputFlag == "json" {
				_MachineSetsList.Items = append(_MachineSetsList.Items, MachineSet)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_MachineSetsList.Items = append(_MachineSetsList.Items, MachineSet)
				continue
			}

			//name
			MachineSetName := MachineSet.Name

			//desired
			desired := ""
			if MachineSet.Spec.Replicas != nil {
				desired = strconv.Itoa(int(*MachineSet.Spec.Replicas))
			}

			//current
			current := strconv.Itoa(int(MachineSet.Status.Replicas))

			//ready
			ready := strconv.Itoa(int(MachineSet.Status.ReadyReplicas))

			//avaialble
			avaialble := strconv.Itoa(int(MachineSet.Status.AvailableReplicas))

			//age
			age := helpers.GetAge(machineYamlPath, MachineSet.GetCreationTimestamp())
			//labels
			labels := helpers.ExtractLabels(MachineSet.GetLabels())
			_list := []string{MachineSet.Namespace, MachineSetName, desired, current, ready, avaialble, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 7, _list)

			if resourceName != "" && resourceName == MachineSetName {
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
			headers = _headers[0:7]
		} else {
			headers = _headers[1:7]
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

	if len(_MachineSetsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _MachineSetsList.Items[0]
	} else {
		resource = _MachineSetsList
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

var MachineSet = &cobra.Command{
	Use:     "machineset",
	Aliases: []string{"machinesets", "machineset.machine.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getMachineSets(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

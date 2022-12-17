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
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	machineapi "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type MachinesItems struct {
	ApiVersion string               `json:"apiVersion"`
	Items      []machineapi.Machine `json:"items"`
}

func getMachines(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "phase", "type", "region", "zone", "age", "node", "providerid", "state"}
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
	var _MachinesList = MachinesItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_machines, err := ioutil.ReadDir(CurrentNamespacePath + "/machine.openshift.io/machines/")
		if err != nil && !allNamespacesFlag {
			fmt.Fprintln(os.Stderr, "No resources found in "+_namespace+" namespace.")
			os.Exit(1)
		}
		for _, f := range _machines {
			machineYamlPath := CurrentNamespacePath + "/machine.openshift.io/machines/" + f.Name()
			_file := helpers.ReadYaml(machineYamlPath)
			Machine := machineapi.Machine{}
			if err := yaml.Unmarshal([]byte(_file), &Machine); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+machineYamlPath)
				os.Exit(1)
			}

			labels := helpers.ExtractLabels(Machine.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}
			if resourceName != "" && resourceName != Machine.Name {
				continue
			}

			if outputFlag == "name" {
				_MachinesList.Items = append(_MachinesList.Items, Machine)
				fmt.Println("machine.machine.openshift.io/" + Machine.Name)
				continue
			}

			if outputFlag == "yaml" {
				_MachinesList.Items = append(_MachinesList.Items, Machine)
				continue
			}

			if outputFlag == "json" {
				_MachinesList.Items = append(_MachinesList.Items, Machine)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_MachinesList.Items = append(_MachinesList.Items, Machine)
				continue
			}

			//name
			MachineName := Machine.Name

			//phase
			phase := ""
			if Machine.Status.Phase != nil {
				phase = *Machine.Status.Phase
			}
			//type
			machineType := helpers.ExtractLabel(Machine.GetLabels(), "machine.openshift.io/instance-type")

			//region:
			region := helpers.ExtractLabel(Machine.GetLabels(), "machine.openshift.io/region")

			//zone
			zone := helpers.ExtractLabel(Machine.GetLabels(), "machine.openshift.io/zone")

			//node
			node := ""
			if Machine.Status.NodeRef != nil {
				node = Machine.Status.NodeRef.Name
			}

			//providerid
			providerid := ""
			if Machine.Spec.ProviderID != nil {
				providerid = *Machine.Spec.ProviderID
			}

			//state
			state := helpers.ExtractLabel(Machine.GetAnnotations(), "machine.openshift.io/instance-state")

			//age
			age := helpers.GetAge(machineYamlPath, Machine.GetCreationTimestamp())

			_list := []string{Machine.Namespace, MachineName, phase, machineType, region, zone, age, node, providerid, state}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 7, _list)

			if resourceName != "" && resourceName == MachineName {
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

	if len(_MachinesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _MachinesList.Items[0]
	} else {
		resource = _MachinesList
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

var Machine = &cobra.Command{
	Use:     "machine",
	Aliases: []string{"machines", "machine.machine.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getMachines(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

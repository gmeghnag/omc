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

package operators

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	"github.com/operator-framework/api/pkg/operators/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

func GetOperatorGroup(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "age"}
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
	var OperatorGroupList = v1.OperatorGroupList{}
	for _, _namespace := range namespaces {
		n_OperatorGroupList := v1.OperatorGroupList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/operators.coreos.com/operatorgroups/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/operators.coreos.com/operatorgroups/" + f.Name()
			_file, err := ioutil.ReadFile(smcpYamlPath)
			if err != nil {
				fmt.Println(err.Error())
			}
			_OperatorGroup := v1.OperatorGroup{}
			if err := yaml.Unmarshal([]byte(_file), &_OperatorGroup); err != nil {
				fmt.Println("Error when trying to unmarshal file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_OperatorGroupList.Items = append(n_OperatorGroupList.Items, _OperatorGroup)
		}
		for _, OperatorGroup := range n_OperatorGroupList.Items {
			labels := helpers.ExtractLabels(OperatorGroup.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}

			if resourceName != "" && resourceName != OperatorGroup.Name {
				continue
			}
			if outputFlag == "name" {
				n_OperatorGroupList.Items = append(n_OperatorGroupList.Items, OperatorGroup)
				fmt.Println("OperatorGroup.operators.coreos.com/" + OperatorGroup.Name)
				continue
			}

			if outputFlag == "yaml" {
				n_OperatorGroupList.Items = append(n_OperatorGroupList.Items, OperatorGroup)
				OperatorGroupList.Items = append(OperatorGroupList.Items, OperatorGroup)
				continue
			}

			if outputFlag == "json" {
				n_OperatorGroupList.Items = append(n_OperatorGroupList.Items, OperatorGroup)
				OperatorGroupList.Items = append(OperatorGroupList.Items, OperatorGroup)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_OperatorGroupList.Items = append(n_OperatorGroupList.Items, OperatorGroup)
				OperatorGroupList.Items = append(OperatorGroupList.Items, OperatorGroup)
				continue
			}

			//name
			OperatorGroupName := OperatorGroup.Name
			//age
			ogYamlPath := fmt.Sprintf("%s/operators.coreos.com/operatorgroups/%s.yaml", CurrentNamespacePath, OperatorGroup.Name)
			age := helpers.GetAge(ogYamlPath, OperatorGroup.GetCreationTimestamp())

			_list := []string{_namespace, OperatorGroupName, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 3, _list)

			if resourceName != "" && resourceName == OperatorGroupName {
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
			headers = _headers[0:3]
		} else {
			headers = _headers[1:3]
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

	if len(OperatorGroupList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = OperatorGroupList.Items[0]
	} else {
		resource = OperatorGroupList
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

var OperatorGroup = &cobra.Command{
	Use:     "operatorgroup",
	Aliases: []string{"og", "operatorgroups", "operatorgroup.operators.coreos.com"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetOperatorGroup(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

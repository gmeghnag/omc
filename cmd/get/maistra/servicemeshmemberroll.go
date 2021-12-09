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
package maistra

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	v1 "maistra.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func GetServiceMeshMemberRoll(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "ready", "status", "age", "members"}

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
	var ServiceMeshMemberRollsList = v1.ServiceMeshMemberRollList{}
	for _, _namespace := range namespaces {
		var n_ServiceMeshMemberRollsList = v1.ServiceMeshMemberRollList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/maistra.io/servicemeshmemberrolls/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/maistra.io/servicemeshmemberrolls/" + f.Name()
			_file := helpers.ReadYaml(smcpYamlPath)
			_ServiceMeshMemberRoll := v1.ServiceMeshMemberRoll{}
			if err := yaml.Unmarshal([]byte(_file), &_ServiceMeshMemberRoll); err != nil {
				fmt.Println("Error when trying to unmarshal file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_ServiceMeshMemberRollsList.Items = append(n_ServiceMeshMemberRollsList.Items, _ServiceMeshMemberRoll)
		}
		for _, ServiceMeshMemberRoll := range n_ServiceMeshMemberRollsList.Items {
			if resourceName != "" && resourceName != ServiceMeshMemberRoll.Name {
				continue
			}

			if outputFlag == "yaml" {
				n_ServiceMeshMemberRollsList.Items = append(n_ServiceMeshMemberRollsList.Items, ServiceMeshMemberRoll)
				ServiceMeshMemberRollsList.Items = append(ServiceMeshMemberRollsList.Items, ServiceMeshMemberRoll)
				continue
			}

			if outputFlag == "json" {
				n_ServiceMeshMemberRollsList.Items = append(n_ServiceMeshMemberRollsList.Items, ServiceMeshMemberRoll)
				ServiceMeshMemberRollsList.Items = append(ServiceMeshMemberRollsList.Items, ServiceMeshMemberRoll)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_ServiceMeshMemberRollsList.Items = append(n_ServiceMeshMemberRollsList.Items, ServiceMeshMemberRoll)
				ServiceMeshMemberRollsList.Items = append(ServiceMeshMemberRollsList.Items, ServiceMeshMemberRoll)
				continue
			}

			//name
			ServiceMeshMemberRollName := ServiceMeshMemberRoll.Name
			//ready
			ready := helpers.ExtractLabel(ServiceMeshMemberRoll.Status.Annotations, "configuredMemberCount")
			//status
			status := ""
			for _, c := range ServiceMeshMemberRoll.Status.Conditions {
				if c.Type == "Ready" {
					status = string(c.Reason)
					break
				}
			}
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/maistra.io/servicemeshmemberrolls/"+ServiceMeshMemberRollName+".yaml", ServiceMeshMemberRoll.GetCreationTimestamp())
			//members
			ServiceMeshMemberRollMembers := ""
			if len(ServiceMeshMemberRoll.Status.Members) != 0 {
				ServiceMeshMemberRollMembers = "[" + strings.Join(ServiceMeshMemberRoll.Status.Members, ", ") + "]"
			}

			labels := helpers.ExtractLabels(ServiceMeshMemberRoll.GetLabels())
			_list := []string{ServiceMeshMemberRoll.Namespace, ServiceMeshMemberRollName, ready, status, age, ServiceMeshMemberRollMembers}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == ServiceMeshMemberRollName {
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
			headers = _headers[0:5]
		} else {
			headers = _headers[1:5]
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

	if len(ServiceMeshMemberRollsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = ServiceMeshMemberRollsList.Items[0]
	} else {
		resource = ServiceMeshMemberRollsList
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

var ServiceMeshMemberRoll = &cobra.Command{
	Use:     "servicemeshmemberroll",
	Aliases: []string{"smmr", "servicemeshmemberrolls"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetServiceMeshMemberRoll(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

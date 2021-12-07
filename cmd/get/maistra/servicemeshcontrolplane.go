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
	v2 "maistra.io/api/core/v2"
	"sigs.k8s.io/yaml"
)

func GetServiceMeshControlPlane(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "ready", "status", "profiles", "version", "age", "image registry"}

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
	var ServiceMeshControlPlanesList = v2.ServiceMeshControlPlaneList{}
	for _, _namespace := range namespaces {
		var n_ServiceMeshControlPlanesList = v2.ServiceMeshControlPlaneList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/maistra.io/servicemeshcontrolplanes/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/maistra.io/servicemeshcontrolplanes/" + f.Name()
			_file := helpers.ReadYaml(smcpYamlPath)
			_ServiceMeshControlPlane := v2.ServiceMeshControlPlane{}
			if err := yaml.Unmarshal([]byte(_file), &_ServiceMeshControlPlane); err != nil {
				fmt.Println("Error when trying to unmarshall file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_ServiceMeshControlPlanesList.Items = append(n_ServiceMeshControlPlanesList.Items, _ServiceMeshControlPlane)
		}
		for _, ServiceMeshControlPlane := range n_ServiceMeshControlPlanesList.Items {
			if resourceName != "" && resourceName != ServiceMeshControlPlane.Name {
				continue
			}

			if outputFlag == "yaml" {
				n_ServiceMeshControlPlanesList.Items = append(n_ServiceMeshControlPlanesList.Items, ServiceMeshControlPlane)
				ServiceMeshControlPlanesList.Items = append(ServiceMeshControlPlanesList.Items, ServiceMeshControlPlane)
				continue
			}

			if outputFlag == "json" {
				n_ServiceMeshControlPlanesList.Items = append(n_ServiceMeshControlPlanesList.Items, ServiceMeshControlPlane)
				ServiceMeshControlPlanesList.Items = append(ServiceMeshControlPlanesList.Items, ServiceMeshControlPlane)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_ServiceMeshControlPlanesList.Items = append(n_ServiceMeshControlPlanesList.Items, ServiceMeshControlPlane)
				ServiceMeshControlPlanesList.Items = append(ServiceMeshControlPlanesList.Items, ServiceMeshControlPlane)
				continue
			}

			//name
			ServiceMeshControlPlaneName := ServiceMeshControlPlane.Name
			//ready
			ready := helpers.ExtractLabel(ServiceMeshControlPlane.Status.Annotations, "readyComponentCount")
			//status
			status := ""
			for _, c := range ServiceMeshControlPlane.Status.Conditions {
				if c.Type == "Ready" {
					status = string(c.Reason)
					break
				}
			}
			//profiles
			profiles := fmt.Sprintf("%q", ServiceMeshControlPlane.Status.AppliedSpec.Profiles)
			//version
			version := ServiceMeshControlPlane.Status.ChartVersion
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/maistra.io/servicemeshcontrolplanes/"+ServiceMeshControlPlaneName+".yaml", ServiceMeshControlPlane.GetCreationTimestamp())
			//image egistry
			registry := ServiceMeshControlPlane.Status.AppliedSpec.Runtime.Defaults.Container.ImageRegistry

			labels := helpers.ExtractLabels(ServiceMeshControlPlane.GetLabels())
			_list := []string{ServiceMeshControlPlane.Namespace, ServiceMeshControlPlaneName, ready, status, profiles, version, age, registry}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == ServiceMeshControlPlaneName {
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

	if len(ServiceMeshControlPlanesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = ServiceMeshControlPlanesList.Items[0]
	} else {
		resource = ServiceMeshControlPlanesList
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

var ServiceMeshControlPlane = &cobra.Command{
	Use:     "servicemeshcontrolplane",
	Aliases: []string{"smcp", "servicemeshcontrolplanes"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetServiceMeshControlPlane(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

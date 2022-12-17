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

type EndpointsItems struct {
	ApiVersion string              `json:"apiVersion"`
	Items      []*corev1.Endpoints `json:"items"`
}

func getEndpoints(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "endpoints", "age"}
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
	var _EndpointsList = EndpointsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items EndpointsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/endpoints.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Fprintln(os.Stderr, "No resources found in "+_namespace+" namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+CurrentNamespacePath+"/core/endpoints.yaml")
			os.Exit(1)
		}

		for _, Endpoint := range _Items.Items {
			labels := helpers.ExtractLabels(Endpoint.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}
			if resourceName != "" && resourceName != Endpoint.Name {
				continue
			}

			if outputFlag == "name" {
				_EndpointsList.Items = append(_EndpointsList.Items, Endpoint)
				fmt.Println("endpoints/" + Endpoint.Name)
				continue
			}

			if outputFlag == "yaml" {
				_EndpointsList.Items = append(_EndpointsList.Items, Endpoint)
				continue
			}

			if outputFlag == "json" {
				_EndpointsList.Items = append(_EndpointsList.Items, Endpoint)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_EndpointsList.Items = append(_EndpointsList.Items, Endpoint)
				continue
			}

			//name
			EndpointName := Endpoint.Name
			if allResources {
				EndpointName = "endpoint/" + EndpointName
			}
			//data
			ep := []string{}
			endpoints := "<none>"
			for _, s := range Endpoint.Subsets {
				if len(s.Addresses) > 0 && len(s.Ports) > 0 {
					for _, a := range s.Addresses {
						for _, p := range s.Ports {
							ep = append(ep, a.IP+":"+strconv.Itoa(int(p.Port)))
						}
					}
				}
			}
			if len(ep) < 4 && len(ep) != 0 {
				endpoints = strings.Join(ep, ",")
			}
			if len(ep) > 3 {
				endpoints = strings.Join(ep[:3], ",") + " + " + strconv.Itoa(len(ep)-3) + " more..."
			}
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/endpoints.yaml", Endpoint.GetCreationTimestamp())
			_list := []string{Endpoint.Namespace, EndpointName, endpoints, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 4, _list)

			if resourceName != "" && resourceName == EndpointName {
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

	if len(_EndpointsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _EndpointsList.Items[0]
	} else {
		resource = _EndpointsList
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
			fmt.Println(_EndpointsList.Items[0].Name)
			resource = _EndpointsList.Items[0]
		} else {
			resource = _EndpointsList
		}
	}
	helpers.PrintOutput(resource, outputFlag, resourceName, allNamespacesFlag, showLabels, _headers, data, jsonPathTemplate)
	*/return false
}

var Endpoint = &cobra.Command{
	Use:     "endpoint",
	Aliases: []string{"endpoints", "ep"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getEndpoints(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

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

	"omc/cmd/helpers"
	"omc/vars"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type PodsItems struct {
	ApiVersion string        `json:"apiVersion"`
	Items      []*corev1.Pod `json:"items"`
}

func getPods(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "ready", "status", "restarts", "age", "ip", "node"}
	var namespaces []string
	if allNamespacesFlag == true {
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	}
	if namespace != "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}
	if namespace == "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}

	var data [][]string
	var _PodsList = PodsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items PodsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/pods.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/pods.yaml")
			os.Exit(1)
		}

		for _, Pod := range _Items.Items {
			// pod path
			if resourceName != "" && resourceName != Pod.Name {
				continue
			}

			if outputFlag == "yaml" {
				_PodsList.Items = append(_PodsList.Items, Pod)
				continue
			}

			if outputFlag == "json" {
				_PodsList.Items = append(_PodsList.Items, Pod)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_PodsList.Items = append(_PodsList.Items, Pod)
				continue
			}

			//name
			PodName := Pod.Name
			if allResources {
				PodName = "pod/" + PodName
			}
			//ContainersReady
			var containers string
			if len(Pod.Spec.Containers) != 0 {
				containers = strconv.Itoa(len(Pod.Spec.Containers))
			} else {
				containers = "0"
			}
			var containerStatuses = Pod.Status.ContainerStatuses

			containers_ready := 0
			for _, i := range containerStatuses {
				if i.Ready == true {
					containers_ready = containers_ready + 1
				}
			}
			ContainersReady := strconv.Itoa(containers_ready) + "/" + containers
			//status
			status := string(Pod.Status.Phase)
			if status == "Succeeded" {
				status = "Completed"
			}
			// restarts
			ContainersRestarts := 0
			for _, i := range containerStatuses {
				if int(i.RestartCount) > ContainersRestarts {
					ContainersRestarts = int(i.RestartCount)
				}
			}
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/pods.yaml", Pod.GetCreationTimestamp())
			//labels
			labels := helpers.ExtractLabels(Pod.GetLabels())
			_list := []string{Pod.Namespace, PodName, ContainersReady, status, strconv.Itoa(ContainersRestarts), age, string(Pod.Status.PodIP), Pod.Spec.NodeName}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 6, _list)

			if resourceName != "" && resourceName == PodName {
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
			headers = _headers[0:6]
		} else {
			headers = _headers[1:6]
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

	if len(_PodsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _PodsList.Items[0]
	} else {
		resource = _PodsList
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
			fmt.Println(_PodsList.Items[0].Name)
			resource = _PodsList.Items[0]
		} else {
			resource = _PodsList
		}
	}
	helpers.PrintOutput(resource, outputFlag, resourceName, allNamespacesFlag, showLabels, _headers, data, jsonPathTemplate)
	*/return false
}

var Pod = &cobra.Command{
	Use:     "pod",
	Aliases: []string{"po", "pods"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getPods(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

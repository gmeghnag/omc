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
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"omc/cmd/helpers"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type PodsItems struct {
	ApiVersion string        `json:"apiVersion"`
	Items      []*corev1.Pod `json:"items"`
}

func getPods(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "ready", "status", "restarts", "age", "ip", "node"}
	// get quay-io-... string
	files, err := ioutil.ReadDir(currentContextPath)
	if err != nil {
		log.Fatal(err)
	}
	var QuayString string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "quay") {
			QuayString = f.Name()
			break
		}
	}
	if QuayString == "" {
		fmt.Println("Some error occurred, wrong must-gather file composition")
		os.Exit(1)
	}
	var namespaces []string
	if allNamespacesFlag == true {
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/" + QuayString + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	}
	if namespace != "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}
	if namespace == "" && !allNamespacesFlag {
		var _namespace = defaultConfigNamespace
		namespaces = append(namespaces, _namespace)
	}

	var data [][]string
	var _PodsList = PodsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items PodsItems
		CurrentNamespacePath := currentContextPath + "/" + QuayString + "/namespaces/" + _namespace
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
			// restarts
			ContainersRestarts := 0
			for _, i := range containerStatuses {
				if int(i.RestartCount) > ContainersRestarts {
					ContainersRestarts = int(i.RestartCount)
				}
			}
			//age
			ResourceFile, _ := os.Stat(CurrentNamespacePath + "/core/pods.yaml")
			t2 := ResourceFile.ModTime()
			t1 := Pod.GetCreationTimestamp()
			diffTime := t2.Sub(t1.Time).String()
			d, _ := time.ParseDuration(diffTime)
			diffTimeString := helpers.FormatDiffTime(d)
			//labels
			labels := helpers.ExtractLabels(Pod.GetLabels())
			_list := []string{Pod.Namespace, PodName, ContainersReady, string(Pod.Status.Phase), strconv.Itoa(ContainersRestarts), diffTimeString, string(Pod.Status.PodIP), Pod.Spec.NodeName}
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
		if allResources {
			return true
		} else {
			fmt.Println("No resources found in " + namespace + " namespace.")
			return true
		}
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

	return false
}

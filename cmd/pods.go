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

func getPods(CurrentContextPath string, DefaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, jsonPathTemplate string) {
	headers := []string{"namespace", "name", "ready", "status", "restarts", "age", "ip", "node"}
	// get quay-io-... string
	files, err := ioutil.ReadDir(CurrentContextPath)
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
		_namespaces, _ := ioutil.ReadDir(CurrentContextPath + "/" + QuayString + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	}
	if namespace != "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}
	if namespace == "" && !allNamespacesFlag {
		var _namespace = DefaultConfigNamespace
		namespaces = append(namespaces, _namespace)
	}

	var data [][]string
	var _PodsList = PodsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items PodsItems
		CurrentNamespacePath := CurrentContextPath + "/" + QuayString + "/namespaces/" + _namespace
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

			var containers string
			if len(Pod.Spec.Containers) != 0 {
				containers = strconv.Itoa(len(Pod.Spec.Containers))
			} else {
				containers = "0"
			}
			var containerStatuses = Pod.Status.ContainerStatuses // DA VALIDARE L'ESISTENZA
			// ready
			containers_ready := 0
			for _, i := range containerStatuses {
				if i.Ready == true {
					containers_ready = containers_ready + 1
				}
			}
			// restarts
			ContainersRestarts := 0
			for _, i := range containerStatuses {
				if int(i.RestartCount) > ContainersRestarts {
					ContainersRestarts = int(i.RestartCount)
				}
			}
			ContainersReady := strconv.Itoa(containers_ready) + "/" + containers
			//age
			PodsFile, err := os.Stat(CurrentNamespacePath + "/core/pods.yaml")

			if err != nil {
				fmt.Println(err)
			}
			// check podfile last time modification as t2
			t2 := PodsFile.ModTime()
			layout := "2006-01-02 15:04:05 -0700 MST"
			t1, _ := time.Parse(layout, Pod.ObjectMeta.CreationTimestamp.String())
			diffTime := t2.Sub(t1).String()
			d, _ := time.ParseDuration(diffTime)
			diffTimeString := helpers.FormatDiffTime(d)
			//return
			_list := []string{Pod.Namespace, Pod.Name, ContainersReady, string(Pod.Status.Phase), strconv.Itoa(ContainersRestarts), diffTimeString, string(Pod.Status.PodIP), Pod.Spec.NodeName}
			if allNamespacesFlag == true {
				if outputFlag == "" {
					data = append(data, _list[0:6]) // -A
				}
				if outputFlag == "wide" {
					data = append(data, _list) // -A -o wide
				}
			} else {
				if outputFlag == "" {
					data = append(data, _list[1:6])
				}
				if outputFlag == "wide" {
					data = append(data, _list[1:]) // -o wide
				}
			}
		}
	}
	if outputFlag == "" {
		if allNamespacesFlag == true {
			helpers.PrintTable(headers[0:6], data) // -A
		} else {
			helpers.PrintTable(headers[1:6], data)
		}
	}
	if outputFlag == "wide" {
		if allNamespacesFlag == true {
			helpers.PrintTable(headers, data) // -A -o wide
		} else {
			helpers.PrintTable(headers[1:], data) // -o wide
		}
	}
	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(_PodsList)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(_PodsList, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(_PodsList, jsonPathTemplate)
	}
}

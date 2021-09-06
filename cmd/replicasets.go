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

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

type ReplicaSetsItems struct {
	ApiVersion string               `json:"apiVersion"`
	Items      []*appsv1.ReplicaSet `json:"items"`
}

func getReplicaSets(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "desired", "current", "ready", "age", "containers", "images", "selector"}

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
	var _ReplicaSetsList = ReplicaSetsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items ReplicaSetsItems
		CurrentNamespacePath := currentContextPath + "/" + QuayString + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/apps/replicasets.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/apps/replicasets.yaml")
			os.Exit(1)
		}

		for _, ReplicaSet := range _Items.Items {
			if resourceName != "" && resourceName != ReplicaSet.Name {
				continue
			}

			if outputFlag == "yaml" {
				_ReplicaSetsList.Items = append(_ReplicaSetsList.Items, ReplicaSet)
				continue
			}

			if outputFlag == "json" {
				_ReplicaSetsList.Items = append(_ReplicaSetsList.Items, ReplicaSet)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_ReplicaSetsList.Items = append(_ReplicaSetsList.Items, ReplicaSet)
				continue
			}

			//name
			ReplicaSetName := ReplicaSet.Name
			if allResources {
				ReplicaSetName = "replicaset/" + ReplicaSetName
			}
			//desired
			desired := strconv.Itoa(int(ReplicaSet.Status.Replicas))
			//current
			current := strconv.Itoa(int(ReplicaSet.Status.AvailableReplicas))
			//ready
			ready := strconv.Itoa(int(ReplicaSet.Status.ReadyReplicas))
			//age
			ResourceFile, _ := os.Stat(CurrentNamespacePath + "/apps/replicasets.yaml")
			t2 := ResourceFile.ModTime()
			t1 := ReplicaSet.GetCreationTimestamp()
			diffTime := t2.Sub(t1.Time).String()
			d, _ := time.ParseDuration(diffTime)
			diffTimeString := helpers.FormatDiffTime(d)
			//containers
			containers := ""
			for _, c := range ReplicaSet.Spec.Template.Spec.Containers {
				containers += fmt.Sprint(c.Name) + ","
			}
			if containers == "" {
				containers = "??"
			} else {
				containers = strings.TrimRight(containers, ",")
			}
			//images
			images := ""
			for _, i := range ReplicaSet.Spec.Template.Spec.Containers {
				images += fmt.Sprint(i.Image) + ","
			}
			if images == "" {
				images = "??"
			} else {
				images = strings.TrimRight(images, ",")
			}
			selector := ""
			for k, v := range ReplicaSet.Spec.Selector.MatchLabels {
				selector += k + "=" + v + ","
			}
			if selector == "" {
				selector = "<none>"
			} else {
				selector = strings.TrimRight(selector, ",")
			}
			//labels
			labels := helpers.ExtractLabels(ReplicaSet.GetLabels())
			_list := []string{ReplicaSet.Namespace, ReplicaSetName, desired, current, ready, diffTimeString, containers, images, selector}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 6, _list)

			if resourceName != "" && resourceName == ReplicaSetName {
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
		resource = _ReplicaSetsList.Items[0]
	} else {
		resource = _ReplicaSetsList
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

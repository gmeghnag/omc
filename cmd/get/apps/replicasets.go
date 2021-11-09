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
package apps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

type ReplicaSetsItems struct {
	ApiVersion string               `json:"apiVersion"`
	Items      []*appsv1.ReplicaSet `json:"items"`
}

func GetReplicaSets(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "desired", "current", "ready", "age", "containers", "images", "selector"}
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
	var _ReplicaSetsList = ReplicaSetsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items ReplicaSetsItems
		CurrentNamespacePath := currentContextPath + "/" + "/namespaces/" + _namespace
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
				ReplicaSetName = "replicaset.apps/" + ReplicaSetName
			}
			//desired
			desired := strconv.Itoa(int(ReplicaSet.Status.Replicas))
			//current
			current := strconv.Itoa(int(ReplicaSet.Status.AvailableReplicas))
			//ready
			ready := strconv.Itoa(int(ReplicaSet.Status.ReadyReplicas))
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/apps/replicasets.yaml", ReplicaSet.GetCreationTimestamp())
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
			_list := []string{ReplicaSet.Namespace, ReplicaSetName, desired, current, ready, age, containers, images, selector}
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

	if len(_ReplicaSetsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
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

var ReplicaSet = &cobra.Command{
	Use:     "replicaset",
	Aliases: []string{"replicasets", "rs", "replicaset.apps"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetReplicaSets(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

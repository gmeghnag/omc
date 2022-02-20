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
	"os"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

type StatefulSetsItems struct {
	ApiVersion string                `json:"apiVersion"`
	Items      []*appsv1.StatefulSet `json:"items"`
}

func GetStatefulSets(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "ready", "age", "containers", "images"}

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
	var _StatefulSetsList = StatefulSetsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items StatefulSetsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/apps/statefulsets.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshal file " + CurrentNamespacePath + "/apps/statefulsets.yaml")
			os.Exit(1)
		}

		for _, StatefulSet := range _Items.Items {
			labels := helpers.ExtractLabels(StatefulSet.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}
			if resourceName != "" && resourceName != StatefulSet.Name {
				continue
			}

			if outputFlag == "name" {
				_StatefulSetsList.Items = append(_StatefulSetsList.Items, StatefulSet)
				fmt.Println("statefulset.apps/" + StatefulSet.Name)
				continue
			}

			if outputFlag == "yaml" {
				_StatefulSetsList.Items = append(_StatefulSetsList.Items, StatefulSet)
				continue
			}

			if outputFlag == "json" {
				_StatefulSetsList.Items = append(_StatefulSetsList.Items, StatefulSet)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_StatefulSetsList.Items = append(_StatefulSetsList.Items, StatefulSet)
				continue
			}

			//name
			StatefulSetName := StatefulSet.Name
			if allResources {
				StatefulSetName = "statefulset.apps/" + StatefulSetName
			}
			//ready
			ready := "??"
			readyReplicas := StatefulSet.Status.ReadyReplicas
			replicas := *StatefulSet.Spec.Replicas
			ready = strconv.Itoa(int(readyReplicas)) + "/" + strconv.Itoa(int(replicas))
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/apps/statefulsets.yaml", StatefulSet.GetCreationTimestamp())
			//containers
			containers := ""
			for _, c := range StatefulSet.Spec.Template.Spec.Containers {
				containers += fmt.Sprint(c.Name) + ","
			}
			if containers == "" {
				containers = "??"
			} else {
				containers = strings.TrimRight(containers, ",")
			}
			//images
			images := ""
			for _, i := range StatefulSet.Spec.Template.Spec.Containers {
				images += fmt.Sprint(i.Image) + ","
			}
			if images == "" {
				images = "??"
			} else {
				images = strings.TrimRight(images, ",")
			}
			_list := []string{StatefulSet.Namespace, StatefulSetName, ready, age, containers, images}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 4, _list)

			if resourceName != "" && resourceName == StatefulSetName {
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

	if len(_StatefulSetsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _StatefulSetsList.Items[0]
	} else {
		resource = _StatefulSetsList
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

var StatefulSet = &cobra.Command{
	Use:     "statefulset",
	Aliases: []string{"statefulsets", "statefulset.apps", "sts"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetStatefulSets(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

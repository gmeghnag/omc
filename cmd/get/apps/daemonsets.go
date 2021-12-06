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

type DaemonsetsItems struct {
	ApiVersion string              `json:"apiVersion"`
	Items      []*appsv1.DaemonSet `json:"items"`
}

func GetDaemonSets(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "desired", "current", "ready", "up-to-date", "available", "node selector", "age", "containers", "images"}

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
	var _DaemonsetsList = DaemonsetsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items DaemonsetsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/apps/daemonsets.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/apps/daemonsets.yaml")
			os.Exit(1)
		}

		for _, Daemonset := range _Items.Items {
			if resourceName != "" && resourceName != Daemonset.Name {
				continue
			}

			if outputFlag == "yaml" {
				_DaemonsetsList.Items = append(_DaemonsetsList.Items, Daemonset)
				continue
			}

			if outputFlag == "json" {
				_DaemonsetsList.Items = append(_DaemonsetsList.Items, Daemonset)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_DaemonsetsList.Items = append(_DaemonsetsList.Items, Daemonset)
				continue
			}

			//name
			DaemonsetName := Daemonset.Name
			if allResources {
				DaemonsetName = "daemonset.apps/" + DaemonsetName
			}
			desired := strconv.Itoa(int(Daemonset.Status.DesiredNumberScheduled))
			//current
			current := strconv.Itoa(int(Daemonset.Status.CurrentNumberScheduled))
			//ready
			ready := strconv.Itoa(int(Daemonset.Status.NumberReady))
			//up-to-date
			upToDate := strconv.Itoa(int(Daemonset.Status.UpdatedNumberScheduled))
			//available
			available := strconv.Itoa(int(Daemonset.Status.NumberAvailable))
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/apps/daemonsets.yaml", Daemonset.GetCreationTimestamp())
			//containers
			containers := ""
			for _, c := range Daemonset.Spec.Template.Spec.Containers {
				containers += fmt.Sprint(c.Name) + ","
			}
			if containers == "" {
				containers = "??"
			} else {
				containers = strings.TrimRight(containers, ",")
			}
			//images
			images := ""
			for _, i := range Daemonset.Spec.Template.Spec.Containers {
				images += fmt.Sprint(i.Image) + ","
			}
			if images == "" {
				images = "??"
			} else {
				images = strings.TrimRight(images, ",")
			}
			//node selector
			selector := ""
			for k, v := range Daemonset.Spec.Template.Spec.NodeSelector {
				selector += k + "=" + v + ","
			}
			if selector == "" {
				selector = "<none>"
			} else {
				selector = strings.TrimRight(selector, ",")
			}
			//labels
			labels := helpers.ExtractLabels(Daemonset.GetLabels())
			_list := []string{Daemonset.Namespace, DaemonsetName, desired, current, ready, upToDate, available, selector, age, containers, images}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 9, _list)

			if resourceName != "" && resourceName == DaemonsetName {
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
			headers = _headers[0:9]
		} else {
			headers = _headers[1:9]
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

	if len(_DaemonsetsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _DaemonsetsList.Items[0]
	} else {
		resource = _DaemonsetsList
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

var DaemonSet = &cobra.Command{
	Use:     "daemonset",
	Aliases: []string{"daemonsets", "daemonset.apps"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetDaemonSets(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

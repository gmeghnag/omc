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
package networking

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type IngressControllerItems struct {
	ApiVersion string                          `json:"apiVersion"`
	Items      []*operatorv1.IngressController `json:"items"`
}

func GetIngressControllers(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "age"}
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
	var IngressControllerList = IngressControllerItems{}
	for _, _namespace := range namespaces {
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_ics, _ := ioutil.ReadDir(CurrentNamespacePath + "/operator.openshift.io/ingresscontrollers/")
		for _, f := range _ics {
			icsYamlPath := CurrentNamespacePath + "/operator.openshift.io/ingresscontrollers/" + f.Name()
			_file, err := ioutil.ReadFile(icsYamlPath)
			if err != nil {
				fmt.Println(err.Error())
			}
			IngressController := &operatorv1.IngressController{}
			if err := yaml.Unmarshal([]byte(_file), &IngressController); err != nil {
				fmt.Println("Error when trying to unmarshal file: " + icsYamlPath)
				os.Exit(1)
			}
			IngressControllerList.Items = append(IngressControllerList.Items, IngressController)

			labels := helpers.ExtractLabels(IngressController.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}

			if resourceName != "" && resourceName != IngressController.Name {
				continue
			}
			if outputFlag == "name" {
				IngressControllerList.Items = append(IngressControllerList.Items, IngressController)
				fmt.Println("ingresscontroller.operator.openshift.io/" + IngressController.Name)
				continue
			}

			if outputFlag == "yaml" {
				IngressControllerList.Items = append(IngressControllerList.Items, IngressController)
				continue
			}

			if outputFlag == "json" {
				IngressControllerList.Items = append(IngressControllerList.Items, IngressController)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				IngressControllerList.Items = append(IngressControllerList.Items, IngressController)
				continue
			}

			age := helpers.GetAge(icsYamlPath, IngressController.GetCreationTimestamp())

			_list := []string{_namespace, IngressController.Name, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 3, _list)

			if resourceName != "" && resourceName == IngressController.Name {
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
			headers = _headers[0:3]
		} else {
			headers = _headers[1:3]
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

	if len(IngressControllerList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = IngressControllerList.Items[0]
	} else {
		resource = IngressControllerList
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

var IngressController = &cobra.Command{
	Use:     "ingresscontroller",
	Aliases: []string{"ingresscontroller.operator.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetIngressControllers(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

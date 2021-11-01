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
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type NamespacesItems struct {
	ApiVersion string             `json:"apiVersion"`
	Items      []corev1.Namespace `json:"items"`
}

func getNamespaces(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	var namespaces []string
	_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
	for _, f := range _namespaces {
		namespaces = append(namespaces, f.Name())
	}

	_headers := []string{"name", "display name", "status"}
	var data [][]string

	_NamespacesList := NamespacesItems{ApiVersion: "v1"}
	for _, f := range namespaces {
		namespaceYamlPath := currentContextPath + "/namespaces/" + f + "/" + f + ".yaml"
		fileExist, _ := helpers.Exists(namespaceYamlPath)
		if !fileExist {
			continue
		}
		_file := helpers.ReadYaml(namespaceYamlPath)
		Namespace := corev1.Namespace{}
		if err := yaml.Unmarshal([]byte(_file), &Namespace); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + namespaceYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != Namespace.Name {
			continue
		}

		if outputFlag == "yaml" {
			_NamespacesList.Items = append(_NamespacesList.Items, Namespace)
			continue
		}

		if outputFlag == "json" {
			_NamespacesList.Items = append(_NamespacesList.Items, Namespace)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_NamespacesList.Items = append(_NamespacesList.Items, Namespace)
			continue
		}

		//displayname
		displayName := ""
		for i, k := range Namespace.ObjectMeta.Annotations {
			if strings.HasPrefix(i, "openshift.io/display-name") {
				displayName = k
				break
			}
		}

		labels := helpers.ExtractLabels(Namespace.GetLabels())
		_list := []string{Namespace.Name, displayName, string(Namespace.Status.Phase)}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 3, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:3] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	var resource interface{}
	if resourceName != "" {
		resource = _NamespacesList.Items[0]
	} else {
		resource = _NamespacesList
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

var Namespace = &cobra.Command{
	Use:     "namespaces",
	Aliases: []string{"ns", "namespace", "project", "projects"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getNamespaces(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

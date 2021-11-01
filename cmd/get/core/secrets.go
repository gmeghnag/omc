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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type SecretsItems struct {
	ApiVersion string           `json:"apiVersion"`
	Items      []*corev1.Secret `json:"items"`
}

func getSecrets(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "type", "data", "age"}
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
	var _SecretsList = SecretsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items SecretsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/secrets.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/secrets.yaml")
			os.Exit(1)
		}

		for _, Secret := range _Items.Items {
			// secret path
			if resourceName != "" && resourceName != Secret.Name {
				continue
			}

			if outputFlag == "yaml" {
				_SecretsList.Items = append(_SecretsList.Items, Secret)
				continue
			}

			if outputFlag == "json" {
				_SecretsList.Items = append(_SecretsList.Items, Secret)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_SecretsList.Items = append(_SecretsList.Items, Secret)
				continue
			}

			//name
			SecretName := Secret.Name
			if allResources {
				SecretName = "secret/" + SecretName
			}
			//type
			secretType := string(Secret.Type)
			//data
			secretData := strconv.Itoa(len(Secret.Data))

			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/secrets.yaml", Secret.GetCreationTimestamp())
			//labels
			labels := helpers.ExtractLabels(Secret.GetLabels())
			_list := []string{Secret.Namespace, SecretName, secretType, secretData, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == SecretName {
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
			headers = _headers[0:5]
		} else {
			headers = _headers[1:5]
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

	if len(_SecretsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _SecretsList.Items[0]
	} else {
		resource = _SecretsList
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

var Secret = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"secrets"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getSecrets(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

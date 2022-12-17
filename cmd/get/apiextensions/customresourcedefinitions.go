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
package apiextensions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

type CustomResourceDefinitionsItems struct {
	ApiVersion string                                     `json:"apiVersion"`
	Items      []apiextensionsv1.CustomResourceDefinition `json:"items"`
}

func getCustomResourceDefinitions(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	customresourcedefinitionsFolderPath := currentContextPath + "/cluster-scoped-resources/apiextensions.k8s.io/customresourcedefinitions/"
	_customresourcedefinitions, _ := ioutil.ReadDir(customresourcedefinitionsFolderPath)

	_headers := []string{"name", "created at"}
	var data [][]string

	_CustomResourceDefinitionsList := CustomResourceDefinitionsItems{ApiVersion: "v1"}
	for _, f := range _customresourcedefinitions {
		customresourcedefinitionYamlPath := customresourcedefinitionsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(customresourcedefinitionYamlPath)
		CustomResourceDefinition := apiextensionsv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal([]byte(_file), &CustomResourceDefinition); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+customresourcedefinitionYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(CustomResourceDefinition.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != CustomResourceDefinition.Name {
			continue
		}

		if outputFlag == "name" {
			_CustomResourceDefinitionsList.Items = append(_CustomResourceDefinitionsList.Items, CustomResourceDefinition)
			fmt.Println("customresourcedefinition.apiextensions.k8s.io/" + CustomResourceDefinition.Name)
			continue
		}

		if outputFlag == "yaml" {
			_CustomResourceDefinitionsList.Items = append(_CustomResourceDefinitionsList.Items, CustomResourceDefinition)
			continue
		}

		if outputFlag == "json" {
			_CustomResourceDefinitionsList.Items = append(_CustomResourceDefinitionsList.Items, CustomResourceDefinition)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_CustomResourceDefinitionsList.Items = append(_CustomResourceDefinitionsList.Items, CustomResourceDefinition)
			continue
		}

		creationTimestamp := CustomResourceDefinition.GetCreationTimestamp()
		createdAt := creationTimestamp.UTC().Format(time.RFC3339Nano)

		_list := []string{CustomResourceDefinition.Name, createdAt}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers // -A
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
		resource = _CustomResourceDefinitionsList.Items[0]
	} else {
		resource = _CustomResourceDefinitionsList
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

var CustomResourceDefinition = &cobra.Command{
	Use:     "customresourcedefinition",
	Aliases: []string{"customresourcedefinitions", "crd", "crds", "customresourcedefinition.apiextensions.k8s.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getCustomResourceDefinitions(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

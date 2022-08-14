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

package operators

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

func GetCatalogSource(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "display", "type", "publisher", "age"}
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
	var CatalogSourceList = v1alpha1.CatalogSourceList{}
	for _, _namespace := range namespaces {
		n_CatalogSourceList := v1alpha1.CatalogSourceList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/operators.coreos.com/catalogsources/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/operators.coreos.com/catalogsources/" + f.Name()
			_file, err := ioutil.ReadFile(smcpYamlPath)
			if err != nil {
				fmt.Println(err.Error())
			}
			_CatalogSource := v1alpha1.CatalogSource{}
			if err := yaml.Unmarshal([]byte(_file), &_CatalogSource); err != nil {
				fmt.Println("Error when trying to unmarshal file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_CatalogSourceList.Items = append(n_CatalogSourceList.Items, _CatalogSource)
		}
		for _, CatalogSource := range n_CatalogSourceList.Items {
			labels := helpers.ExtractLabels(CatalogSource.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}

			if resourceName != "" && resourceName != CatalogSource.Name {
				continue
			}
			if outputFlag == "name" {
				n_CatalogSourceList.Items = append(n_CatalogSourceList.Items, CatalogSource)
				fmt.Println("catalogsource.operators.coreos.com/" + CatalogSource.Name)
				continue
			}

			if outputFlag == "yaml" {
				n_CatalogSourceList.Items = append(n_CatalogSourceList.Items, CatalogSource)
				CatalogSourceList.Items = append(CatalogSourceList.Items, CatalogSource)
				continue
			}

			if outputFlag == "json" {
				n_CatalogSourceList.Items = append(n_CatalogSourceList.Items, CatalogSource)
				CatalogSourceList.Items = append(CatalogSourceList.Items, CatalogSource)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_CatalogSourceList.Items = append(n_CatalogSourceList.Items, CatalogSource)
				CatalogSourceList.Items = append(CatalogSourceList.Items, CatalogSource)
				continue
			}

			//name
			CatalogSourceName := CatalogSource.Name
			//display name
			displayName := CatalogSource.Spec.DisplayName
			//type
			csType := string(CatalogSource.Spec.SourceType)
			//publisher
			publisher := CatalogSource.Spec.Publisher
			//age
			csYamlPath := fmt.Sprintf("%s/operators.coreos.com/catalogsources/%s.yaml", CurrentNamespacePath, CatalogSource.Name)
			age := helpers.GetAge(csYamlPath, CatalogSource.GetCreationTimestamp())

			_list := []string{_namespace, CatalogSourceName, displayName, csType, publisher, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 6, _list)

			if resourceName != "" && resourceName == CatalogSourceName {
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

	if len(CatalogSourceList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = CatalogSourceList.Items[0]
	} else {
		resource = CatalogSourceList
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

var CatalogSource = &cobra.Command{
	Use:     "catalogsource",
	Aliases: []string{"catsrc", "catalogsources", "catalogsource.operators.coreos.com"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetCatalogSource(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

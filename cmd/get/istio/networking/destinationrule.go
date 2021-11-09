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
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strings"

	"github.com/spf13/cobra"
	v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/yaml"
)

func GetDestinationRule(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name"}

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
	var DestinationRulesList = v1alpha3.DestinationRuleList{}
	for _, _namespace := range namespaces {
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/networking.io/destinationrules/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/networking.io/destinationrules/" + f.Name()
			_file := helpers.ReadYaml(smcpYamlPath)
			_DestinationRule := v1alpha3.DestinationRule{}
			if err := yaml.Unmarshal([]byte(_file), &_DestinationRule); err != nil {
				fmt.Println("Error when trying to unmarshall file: " + smcpYamlPath)
				os.Exit(1)
			}
			fmt.Println("ddddd")
			DestinationRulesList.Items = append(DestinationRulesList.Items, _DestinationRule)
		}
		for _, DestRule := range DestinationRulesList.Items {
			if resourceName != "" && resourceName != DestRule.Name {
				continue
			}

			if outputFlag == "yaml" {
				DestinationRulesList.Items = append(DestinationRulesList.Items, DestRule)
				continue
			}

			if outputFlag == "json" {
				DestinationRulesList.Items = append(DestinationRulesList.Items, DestRule)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				DestinationRulesList.Items = append(DestinationRulesList.Items, DestRule)
				continue
			}

			//name
			DestinationRuleName := DestRule.Name

			labels := helpers.ExtractLabels(DestRule.GetLabels())
			_list := []string{DestRule.Namespace, DestinationRuleName}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == DestinationRuleName {
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

	if len(DestinationRulesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = DestinationRulesList.Items[0]
	} else {
		resource = DestinationRulesList
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

var DestinationRule = &cobra.Command{
	Use:     "destinationrule",
	Aliases: []string{"dr", "destinationrules"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetDestinationRule(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

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
package apiregistration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/yaml"
)

type APIServiceList struct {
	ApiVersion string                         `json:"apiVersion"`
	Items      []apiregistrationv1.APIService `json:"items"`
	Kind       string                         `json:"kind"`
}

func getApiService(currentContextPath string, namespace string, resourcesNames []string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	apiservicesFolderPath := currentContextPath + "/cluster-scoped-resources/apiregistration.k8s.io/apiservices/"
	_apiservices, _ := ioutil.ReadDir(apiservicesFolderPath)

	_headers := []string{"name", "service", "available", "age"}
	var data [][]string

	_ApiServicesList := APIServiceList{ApiVersion: "v1", Kind: "List"}
	for _, f := range _apiservices {
		apiserviceYamlPath := apiservicesFolderPath + f.Name()
		_file := helpers.ReadYaml(apiserviceYamlPath)
		ApiService := apiregistrationv1.APIService{}
		if err := yaml.Unmarshal([]byte(_file), &ApiService); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+apiserviceYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(ApiService.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if len(resourcesNames) != 0 && !helpers.StringInSlice(ApiService.Name, resourcesNames) {
			continue
		}

		if outputFlag == "name" {
			_ApiServicesList.Items = append(_ApiServicesList.Items, ApiService)
			fmt.Println("apiservice.apiregistration.k8s.io/" + ApiService.Name)
			continue
		}

		if outputFlag == "yaml" {
			_ApiServicesList.Items = append(_ApiServicesList.Items, ApiService)
			continue
		}

		if outputFlag == "json" {
			_ApiServicesList.Items = append(_ApiServicesList.Items, ApiService)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ApiServicesList.Items = append(_ApiServicesList.Items, ApiService)
			continue
		}

		// SERVICE
		service := "Local"
		if ApiService.Spec.Service != nil {
			service = ApiService.Spec.Service.Namespace + "/" + ApiService.Spec.Service.Name
		}

		// AVAILABLE
		available := "Unknown"
		for _, condition := range ApiService.Status.Conditions {
			if condition.Type == "Available" {
				available = string(condition.Status)
				if available != "True" {
					available = string(condition.Status) + " (" + condition.Reason + ")"
				}
				break
			}
		}

		//AGE
		age := helpers.GetAge(apiserviceYamlPath, ApiService.GetCreationTimestamp())

		_list := []string{ApiService.Name, service, available, age}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 4, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:4] // -A
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
	resource = _ApiServicesList
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

var ApiService = &cobra.Command{
	Use:     "apiservices",
	Aliases: []string{"apiservice", "apiservice.apiregistration.k8s.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getApiService(vars.MustGatherRootPath, vars.Namespace, args, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

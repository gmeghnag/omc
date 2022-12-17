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
package apiserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	v1 "github.com/openshift/api/apiserver/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

type ApiRequestCountsItems struct {
	ApiVersion string               `json:"apiVersion"`
	Items      []v1.APIRequestCount `json:"items"`
}

func getAPIRequestCount(currentContextPath string, namespace string, resourcesNames []string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	apirequestcountsFolderPath := currentContextPath + "/cluster-scoped-resources/apiserver.openshift.io/apirequestcounts/"
	_apirequestcounts, _ := ioutil.ReadDir(apirequestcountsFolderPath)

	_headers := []string{"name", "removedinrelease", "requestsincurrenthour", "requestsinlast24h"}
	var data [][]string

	_ApiRequestCountsList := ApiRequestCountsItems{ApiVersion: "v1"}
	for _, f := range _apirequestcounts {
		apirequestcountYamlPath := apirequestcountsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(apirequestcountYamlPath)
		APIRequestCount := v1.APIRequestCount{}
		if err := yaml.Unmarshal([]byte(_file), &APIRequestCount); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+apirequestcountYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(APIRequestCount.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if len(resourcesNames) != 0 && !helpers.StringInSlice(APIRequestCount.Name, resourcesNames) {
			continue
		}

		if outputFlag == "name" {
			_ApiRequestCountsList.Items = append(_ApiRequestCountsList.Items, APIRequestCount)
			fmt.Println("apirequestcount.apiserver.openshift.io/" + APIRequestCount.Name)
			continue
		}

		if outputFlag == "yaml" {
			_ApiRequestCountsList.Items = append(_ApiRequestCountsList.Items, APIRequestCount)
			continue
		}

		if outputFlag == "json" {
			_ApiRequestCountsList.Items = append(_ApiRequestCountsList.Items, APIRequestCount)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ApiRequestCountsList.Items = append(_ApiRequestCountsList.Items, APIRequestCount)
			continue
		}
		//Name
		ApiName := APIRequestCount.Name

		//removedinrelease
		removedInRelease := APIRequestCount.Status.RemovedInRelease

		// requestincurrenthour
		requestInCurrentHour := APIRequestCount.Status.CurrentHour.RequestCount

		//requestinlast24h
		requestInLast24H := 0
		for _, k := range APIRequestCount.Status.Last24h {
			requestInLast24H += int(k.RequestCount)
		}

		_list := []string{ApiName, removedInRelease, strconv.Itoa(int(requestInCurrentHour)), strconv.Itoa(requestInLast24H)}
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
	resource = _ApiRequestCountsList
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

var APIRequestCount = &cobra.Command{
	Use:     "apirequestcount",
	Aliases: []string{"apirequestcounts", "apirequestcount.apiserver.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourcesNames := args
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getAPIRequestCount(vars.MustGatherRootPath, vars.Namespace, resourcesNames, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

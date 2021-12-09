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
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

func getProxies(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	proxiesFolderPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/proxies/"
	_proxies, _ := ioutil.ReadDir(proxiesFolderPath)

	_headers := []string{"name", "age"}
	var data [][]string

	_ProxiesList := configv1.ProxyList{}
	for _, f := range _proxies {
		proxyYamlPath := proxiesFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(proxyYamlPath)
		Proxy := configv1.Proxy{}
		if err := yaml.Unmarshal([]byte(_file), &Proxy); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + proxyYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != Proxy.Name {
			continue
		}

		if outputFlag == "yaml" {
			_ProxiesList.Items = append(_ProxiesList.Items, Proxy)
			continue
		}

		if outputFlag == "json" {
			_ProxiesList.Items = append(_ProxiesList.Items, Proxy)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ProxiesList.Items = append(_ProxiesList.Items, Proxy)
			continue
		}
		//Name
		proxyName := Proxy.Name
		age := helpers.GetAge(proxyYamlPath, Proxy.GetCreationTimestamp())

		labels := helpers.ExtractLabels(Proxy.GetLabels())
		_list := []string{proxyName, age}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:2] // -A
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
		resource = _ProxiesList.Items[0]
	} else {
		resource = _ProxiesList
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

var Proxy = &cobra.Command{
	Use:     "proxy",
	Aliases: []string{"proxies"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getProxies(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

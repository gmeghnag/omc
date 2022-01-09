/*
Copyright © 2021 Christian Passarelli <cpassare@redhat.com>

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
package operator

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

func getAuthentications(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	authenticationsFolderPath := currentContextPath + "/cluster-scoped-resources/operator.openshift.io/authentications/"
	_authentications, _ := ioutil.ReadDir(authenticationsFolderPath)

	_headers := []string{"name", "age"}
	var data [][]string

	_AuthenticationsList := operatorv1.AuthenticationList{}
	for _, f := range _authentications {
		authenticationsYamlPath := authenticationsFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(authenticationsYamlPath)
		Authentication := operatorv1.Authentication{}
		if err := yaml.Unmarshal([]byte(_file), &Authentication); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + authenticationsYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != Authentication.Name {
			continue
		}

		if outputFlag == "yaml" {
			_AuthenticationsList.Items = append(_AuthenticationsList.Items, Authentication)
			continue
		}

		if outputFlag == "json" {
			_AuthenticationsList.Items = append(_AuthenticationsList.Items, Authentication)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_AuthenticationsList.Items = append(_AuthenticationsList.Items, Authentication)
			continue
		}
		//Name
		authenticationName := Authentication.Name
		age := helpers.GetAge(authenticationsYamlPath, Authentication.GetCreationTimestamp())

		labels := helpers.ExtractLabels(Authentication.GetLabels())
		_list := []string{authenticationName, age}
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
		resource = _AuthenticationsList.Items[0]
	} else {
		resource = _AuthenticationsList
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

var Authentication = &cobra.Command{
	Use:     "authentications.operator",
	Aliases: []string{"authentication.operator", "authentication.operator.openshift.io", "authentications.operator", "authentications.operator.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getAuthentications(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

/*
Copyright © 2021 bverschueren@redhat.com

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
package baremetalhost

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	metal3v1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func getBareMetalHosts(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
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
	var bareMetalHostsList = &metal3v1alpha1.BareMetalHostList{}
	for _, _namespace := range namespaces {
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_baremetalhosts, err := ioutil.ReadDir(CurrentNamespacePath + "/metal3.io/baremetalhosts/")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		for _, f := range _baremetalhosts {
			bmhYamlPath := CurrentNamespacePath + "/metal3.io/baremetalhosts/" + f.Name()
			_file := helpers.ReadYaml(bmhYamlPath)
			BareMetalHost := &metal3v1alpha1.BareMetalHost{}
			if err := yaml.Unmarshal([]byte(_file), &BareMetalHost); err != nil {
				fmt.Println("Error when trying to unmarshal file " + bmhYamlPath)
				os.Exit(1)
			}

			labels := helpers.ExtractLabels(BareMetalHost.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}
			if resourceName != "" && resourceName != BareMetalHost.Name {
				continue
			}

			bareMetalHostsList.Items = append(bareMetalHostsList.Items, *BareMetalHost)

			//age
			age := helpers.GetAge(bmhYamlPath, BareMetalHost.GetCreationTimestamp())

			_list := []string{BareMetalHost.Namespace, BareMetalHost.Name, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 3, _list)

			if resourceName != "" && resourceName == BareMetalHost.Name {
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

	if len(bareMetalHostsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = bareMetalHostsList.Items[0]
	} else {
		resource = bareMetalHostsList
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

var BareMetalHost = &cobra.Command{
	Use:     "bareemetalhost",
	Aliases: []string{"baremetalhost", "bmh", "baremetalhost.metal3.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getBareMetalHosts(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

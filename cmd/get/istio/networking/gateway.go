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
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	v1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/yaml"
)

func GetGateway(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
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
	var GatewaysList = v1beta1.GatewayList{}
	for _, _namespace := range namespaces {
		n_GatewaysList := v1beta1.GatewayList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/networking.istio.io/gateways/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/networking.istio.io/gateways/" + f.Name()
			_file := helpers.ReadYaml(smcpYamlPath)
			_Gateway := v1beta1.Gateway{}
			if err := yaml.Unmarshal([]byte(_file), &_Gateway); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+smcpYamlPath)
				os.Exit(1)
			}
			n_GatewaysList.Items = append(n_GatewaysList.Items, _Gateway)
		}
		for _, _Gateway := range n_GatewaysList.Items {
			labels := helpers.ExtractLabels(_Gateway.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}

			if resourceName != "" && resourceName != _Gateway.Name {
				continue
			}
			if outputFlag == "name" {
				n_GatewaysList.Items = append(n_GatewaysList.Items, _Gateway)
				fmt.Println("gateway.networking.istio.io/" + _Gateway.Name)
				continue
			}

			if outputFlag == "yaml" {
				n_GatewaysList.Items = append(n_GatewaysList.Items, _Gateway)
				GatewaysList.Items = append(GatewaysList.Items, _Gateway)
				continue
			}

			if outputFlag == "json" {
				n_GatewaysList.Items = append(n_GatewaysList.Items, _Gateway)
				GatewaysList.Items = append(GatewaysList.Items, _Gateway)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_GatewaysList.Items = append(n_GatewaysList.Items, _Gateway)
				GatewaysList.Items = append(GatewaysList.Items, _Gateway)
				continue
			}

			//name
			GatewayName := _Gateway.Name

			age := helpers.GetAge(CurrentNamespacePath+"/networking.istio.io/gateways/"+GatewayName+".yaml", _Gateway.GetCreationTimestamp())

			_list := []string{_Gateway.Namespace, GatewayName, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 3, _list)

			if resourceName != "" && resourceName == GatewayName {
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

	if len(GatewaysList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = GatewaysList.Items[0]
	} else {
		resource = GatewaysList
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

var Gateway = &cobra.Command{
	Use:     "gateway",
	Aliases: []string{"gw", "gateways", "gateway.networking.istio.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetGateway(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

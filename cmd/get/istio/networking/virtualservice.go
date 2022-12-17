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

func GetVirtualService(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "gateways", "hosts", "age"}

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
	var VirtualServicesList = v1beta1.VirtualServiceList{}
	for _, _namespace := range namespaces {
		n_VirtualServicesList := v1beta1.VirtualServiceList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/networking.istio.io/virtualservices/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/networking.istio.io/virtualservices/" + f.Name()
			_file := helpers.ReadYaml(smcpYamlPath)
			_VirtualService := v1beta1.VirtualService{}
			if err := yaml.Unmarshal([]byte(_file), &_VirtualService); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+smcpYamlPath)
				os.Exit(1)
			}
			n_VirtualServicesList.Items = append(n_VirtualServicesList.Items, _VirtualService)
		}
		for _, VService := range n_VirtualServicesList.Items {
			labels := helpers.ExtractLabels(VService.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}

			if resourceName != "" && resourceName != VService.Name {
				continue
			}
			if outputFlag == "name" {
				n_VirtualServicesList.Items = append(n_VirtualServicesList.Items, VService)
				fmt.Println("virtualservice.networking.istio.io/" + VService.Name)
				continue
			}

			if outputFlag == "yaml" {
				n_VirtualServicesList.Items = append(n_VirtualServicesList.Items, VService)
				VirtualServicesList.Items = append(VirtualServicesList.Items, VService)
				continue
			}

			if outputFlag == "json" {
				n_VirtualServicesList.Items = append(n_VirtualServicesList.Items, VService)
				VirtualServicesList.Items = append(VirtualServicesList.Items, VService)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_VirtualServicesList.Items = append(n_VirtualServicesList.Items, VService)
				VirtualServicesList.Items = append(VirtualServicesList.Items, VService)
				continue
			}

			//name
			VirtualServiceName := VService.Name

			//gateways
			VirtualServiceGateways := ""
			if len(VService.Spec.Gateways) != 0 {
				VirtualServiceGateways = "[" + strings.Join(VService.Spec.Gateways, ", ") + "]"
			}
			//hosts
			VirtualServiceHosts := ""
			if len(VService.Spec.Hosts) != 0 {
				VirtualServiceHosts = "[" + strings.Join(VService.Spec.Hosts, ", ") + "]"
			}

			age := helpers.GetAge(CurrentNamespacePath+"/networking.istio.io/virtualservices/"+VirtualServiceName+".yaml", VService.GetCreationTimestamp())

			_list := []string{VService.Namespace, VirtualServiceName, VirtualServiceGateways, VirtualServiceHosts, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == VirtualServiceName {
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

	if len(VirtualServicesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = VirtualServicesList.Items[0]
	} else {
		resource = VirtualServicesList
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

var VirtualService = &cobra.Command{
	Use:     "virtualservice",
	Aliases: []string{"vs", "virtualservices", "virtualservice.networking.istio.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetVirtualService(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

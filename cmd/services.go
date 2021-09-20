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
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type ServicesItems struct {
	ApiVersion string            `json:"apiVersion"`
	Items      []*corev1.Service `json:"items"`
}

func getServices(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "type", "cluster-ip", "external-ip", "port(s)", "age", "selector"}
	var namespaces []string
	if allNamespacesFlag == true {
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	}
	if namespace != "" && !allNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}
	if namespace == "" && !allNamespacesFlag {
		var _namespace = defaultConfigNamespace
		namespaces = append(namespaces, _namespace)
	}

	var data [][]string
	var _ServicesList = ServicesItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items ServicesItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/services.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/services.yaml")
			os.Exit(1)
		}

		for _, Service := range _Items.Items {
			if resourceName != "" && resourceName != Service.Name {
				continue
			}

			if outputFlag == "yaml" {
				_ServicesList.Items = append(_ServicesList.Items, Service)
				continue
			}

			if outputFlag == "json" {
				_ServicesList.Items = append(_ServicesList.Items, Service)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_ServicesList.Items = append(_ServicesList.Items, Service)
				continue
			}

			//name
			ServiceName := Service.Name
			if allResources {
				ServiceName = "service/" + ServiceName
			}

			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/services.yaml", Service.GetCreationTimestamp())

			//cluster-ip
			ClusterIp := "<none>"
			if Service.Spec.ClusterIP != "" {
				ClusterIp = Service.Spec.ClusterIP
			}
			//external-ip
			externalIp := "<none>"
			if string(Service.Spec.Type) == "ExternalName" {
				externalIp = Service.Spec.ExternalName
			}
			if string(Service.Spec.Type) == "ClusterIp" {
				externalIp = Service.Spec.ClusterIP
			}
			if string(Service.Spec.Type) == "LoadBalancer" {
				externalIp = ""
				for _, p := range Service.Status.LoadBalancer.Ingress {
					externalIp += fmt.Sprint(p.Hostname) + ","
				}
				externalIp = strings.TrimRight(externalIp, ",")
			}
			//ports
			ports := ""
			for _, p := range Service.Spec.Ports {
				ports += fmt.Sprint(p.Port) + "/" + string(p.Protocol) + ","
			}
			if ports == "" {
				ports = "<none>"
			} else {
				ports = strings.TrimRight(ports, ",")
			}
			//selector
			selector := ""
			for k, v := range Service.Spec.Selector {
				selector += k + "=" + v + ","
			}
			if selector == "" {
				selector = "<none>"
			} else {
				selector = strings.TrimRight(selector, ",")
			}
			//labels
			labels := helpers.ExtractLabels(Service.GetLabels())
			_list := []string{Service.Namespace, ServiceName, string(Service.Spec.Type), ClusterIp, externalIp, ports, age, selector}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 7, _list)

			if resourceName != "" && resourceName == ServiceName {
				break
			}
		}
		if namespace != "" && _namespace == namespace {
			break
		}
	}

	if (outputFlag == "" || outputFlag == "wide") && len(data) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + defaultConfigNamespace + " namespace.")
		}
		return true
	}

	var headers []string
	if outputFlag == "" {
		if allNamespacesFlag == true {
			headers = _headers[0:7]
		} else {
			headers = _headers[1:7]
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

	if len(_ServicesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + defaultConfigNamespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _ServicesList.Items[0]
	} else {
		resource = _ServicesList
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

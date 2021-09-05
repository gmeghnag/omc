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
	"log"
	"omc/cmd/helpers"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type ServicesItems struct {
	ApiVersion string            `json:"apiVersion"`
	Items      []*corev1.Service `json:"items"`
}

func getServices(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, jsonPathTemplate string, allResources bool) {
	headers := []string{"namespace", "name", "type", "cluster-ip", "external-ip", "port(s)", "age", "selector"}
	// get quay-io-... string
	files, err := ioutil.ReadDir(currentContextPath)
	if err != nil {
		log.Fatal(err)
	}
	var QuayString string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "quay") {
			QuayString = f.Name()
			break
		}
	}
	if QuayString == "" {
		fmt.Println("Some error occurred, wrong must-gather file composition")
		os.Exit(1)
	}
	var namespaces []string
	if allNamespacesFlag == true {
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/" + QuayString + "/namespaces/")
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
		CurrentNamespacePath := currentContextPath + "/" + QuayString + "/namespaces/" + _namespace
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
			ServicesFile, _ := os.Stat(CurrentNamespacePath + "/core/services.yaml")

			t2 := ServicesFile.ModTime()
			layout := "2006-01-02 15:04:05 -0700 MST"
			t1, _ := time.Parse(layout, Service.ObjectMeta.CreationTimestamp.String())
			diffTime := t2.Sub(t1).String()
			d, _ := time.ParseDuration(diffTime)
			diffTimeString := helpers.FormatDiffTime(d)
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
				externalIp = Service.Spec.LoadBalancerIP
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
			_list := []string{Service.Namespace, ServiceName, string(Service.Spec.Type), ClusterIp, externalIp, ports, diffTimeString, selector}
			if allNamespacesFlag == true {
				if outputFlag == "" {
					data = append(data, _list[0:7]) // -A
				}
				if outputFlag == "wide" {
					data = append(data, _list) // -A -o wide
				}
			} else {
				if outputFlag == "" {
					data = append(data, _list[1:7])
				}
				if outputFlag == "wide" {
					data = append(data, _list[1:]) // -o wide
				}
			}
		}
	}
	if outputFlag == "" {
		if allNamespacesFlag == true {
			helpers.PrintTable(headers[0:7], data) // -A
		} else {
			helpers.PrintTable(headers[1:7], data)
		}
	}
	if outputFlag == "wide" {
		if allNamespacesFlag == true {
			helpers.PrintTable(headers, data) // -A -o wide
		} else {
			helpers.PrintTable(headers[1:], data) // -o wide
		}
	}
	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(_ServicesList)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(_ServicesList, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(_ServicesList, jsonPathTemplate)
	}
}

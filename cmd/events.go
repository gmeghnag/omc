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

type EventsItems struct {
	ApiVersion string          `json:"apiVersion"`
	Items      []*corev1.Event `json:"items"`
}

func getEvents(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "last seen", "type", "reason", "object", "message"}
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
	var _EventsList = EventsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items EventsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/events.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/events.yaml")
			os.Exit(1)
		}

		for _, Event := range _Items.Items {
			if resourceName != "" && resourceName != Event.Name {
				continue
			}

			if outputFlag == "yaml" {
				_EventsList.Items = append(_EventsList.Items, Event)
				continue
			}

			if outputFlag == "json" {
				_EventsList.Items = append(_EventsList.Items, Event)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_EventsList.Items = append(_EventsList.Items, Event)
				continue
			}

			//last seen
			lastSeenDiffTimeString := helpers.GetAge(CurrentNamespacePath+"/core/events.yaml", Event.LastTimestamp)

			//type
			eventType := Event.Type
			//reason
			reason := Event.Reason
			//object
			object := strings.ToLower(Event.InvolvedObject.Kind) + "/" + Event.InvolvedObject.Name
			//message
			message := Event.Message
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/events.yaml", Event.GetCreationTimestamp())
			//containers

			//labels
			labels := helpers.ExtractLabels(Event.GetLabels())
			_list := []string{Event.Namespace, lastSeenDiffTimeString, eventType, reason, object, message, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 6, _list)

			if resourceName != "" && resourceName == Event.Name {
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
			headers = _headers[0:6]
		} else {
			headers = _headers[1:6]
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
	var resource interface{}
	if resourceName != "" {
		resource = _EventsList.Items[0]
	} else {
		resource = _EventsList
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

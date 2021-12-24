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
package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type EventsItems struct {
	ApiVersion string          `json:"apiVersion"`
	Items      []*corev1.Event `json:"items"`
}

func getEvents(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "last seen", "type", "reason", "object", "message"}
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
			fmt.Println("Error when trying to unmarshal file " + CurrentNamespacePath + "/core/events.yaml")
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
			fmt.Println("No resources found in " + namespace + " namespace.")
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

	if len(_EventsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
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

var Event = &cobra.Command{
	Use:     "event",
	Aliases: []string{"events", "ev"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getEvents(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

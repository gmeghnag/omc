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
	"strconv"
	"strings"

	v1 "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/yaml"
)

type DeploymentConfigsItems struct {
	ApiVersion string                 `json:"apiVersion"`
	Items      []*v1.DeploymentConfig `json:"items"`
}

func getDeploymentConfigs(currentContextPath string, defaultConfigNamespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "revision", "desired", "current", "triggered by"}

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
	var _DeploymentConfigsList = DeploymentConfigsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items DeploymentConfigsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/apps.openshift.io/deploymentconfigs.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/apps.openshift.io/deploymentconfigs.yaml")
			os.Exit(1)
		}

		for _, DeploymentConfig := range _Items.Items {
			if resourceName != "" && resourceName != DeploymentConfig.Name {
				continue
			}

			if outputFlag == "yaml" {
				_DeploymentConfigsList.Items = append(_DeploymentConfigsList.Items, DeploymentConfig)
				continue
			}

			if outputFlag == "json" {
				_DeploymentConfigsList.Items = append(_DeploymentConfigsList.Items, DeploymentConfig)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_DeploymentConfigsList.Items = append(_DeploymentConfigsList.Items, DeploymentConfig)
				continue
			}

			//name
			DeploymentConfigName := DeploymentConfig.Name
			if allResources {
				DeploymentConfigName = "deploymentconfig.apps.openshift.io/" + DeploymentConfigName
			}
			//revision
			revision := strconv.Itoa(int(DeploymentConfig.Status.LatestVersion))
			//desiredReplicas
			desiredReplicas := strconv.Itoa(int(DeploymentConfig.Spec.Replicas))
			//current
			currentReplicas := strconv.Itoa(int(DeploymentConfig.Status.ReadyReplicas))
			//triggered by
			triggeredBy := ""
			triggers := DeploymentConfig.Spec.Triggers
			for _, k := range triggers {
				if k.Type == "ConfigChange" {
					triggeredBy += "config,"
				}
				if k.Type == "ImageChange" {
					triggeredBy += "image" + "(" + k.ImageChangeParams.From.Name + "),"
				}
			}
			triggeredBy = strings.TrimRight(triggeredBy, ",")
			//labels
			labels := helpers.ExtractLabels(DeploymentConfig.GetLabels())
			_list := []string{DeploymentConfig.Namespace, DeploymentConfigName, revision, desiredReplicas, currentReplicas, triggeredBy}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 6, _list)

			if resourceName != "" && resourceName == DeploymentConfigName {
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

	if len(_DeploymentConfigsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + defaultConfigNamespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _DeploymentConfigsList.Items[0]
	} else {
		resource = _DeploymentConfigsList
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

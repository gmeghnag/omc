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
package apps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

type DeploymentsItems struct {
	ApiVersion string               `json:"apiVersion"`
	Items      []*appsv1.Deployment `json:"items"`
}

func GetDeployments(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "ready", "up-to-date", "available", "age", "containers", "images"}

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
	var _DeploymentsList = DeploymentsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items DeploymentsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/apps/deployments.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/apps/deployments.yaml")
			os.Exit(1)
		}

		for _, Deployment := range _Items.Items {
			if resourceName != "" && resourceName != Deployment.Name {
				continue
			}

			if outputFlag == "yaml" {
				_DeploymentsList.Items = append(_DeploymentsList.Items, Deployment)
				continue
			}

			if outputFlag == "json" {
				_DeploymentsList.Items = append(_DeploymentsList.Items, Deployment)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_DeploymentsList.Items = append(_DeploymentsList.Items, Deployment)
				continue
			}

			//name
			DeploymentName := Deployment.Name
			if allResources {
				DeploymentName = "deployment.apps/" + DeploymentName
			}
			//ready
			ready := "??"
			readyReplicas := Deployment.Status.ReadyReplicas
			replicas := *Deployment.Spec.Replicas
			ready = strconv.Itoa(int(readyReplicas)) + "/" + strconv.Itoa(int(replicas))
			//up-to-date
			upToDateReplicas := "??"
			upToDateReplicas = strconv.Itoa(int(Deployment.Status.UpdatedReplicas))
			//available
			availableReplicas := "??"
			availableReplicas = strconv.Itoa(int(Deployment.Status.AvailableReplicas))
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/apps/deployments.yaml", Deployment.GetCreationTimestamp())
			//containers
			containers := ""
			for _, c := range Deployment.Spec.Template.Spec.Containers {
				containers += fmt.Sprint(c.Name) + ","
			}
			if containers == "" {
				containers = "??"
			} else {
				containers = strings.TrimRight(containers, ",")
			}
			//images
			images := ""
			for _, i := range Deployment.Spec.Template.Spec.Containers {
				images += fmt.Sprint(i.Image) + ","
			}
			if images == "" {
				images = "??"
			} else {
				images = strings.TrimRight(images, ",")
			}
			//labels
			labels := helpers.ExtractLabels(Deployment.GetLabels())
			_list := []string{Deployment.Namespace, DeploymentName, ready, upToDateReplicas, availableReplicas, age, containers, images}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == DeploymentName {
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

	if len(_DeploymentsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _DeploymentsList.Items[0]
	} else {
		resource = _DeploymentsList
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

var Deployment = &cobra.Command{
	Use:     "deployment",
	Aliases: []string{"deployments", "deployment.apps"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetDeployments(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

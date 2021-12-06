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
package operators

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	v1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func GetClusterServiceVersion(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "display", "version", "replaces", "phase"}
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
	var ClusterServiceVersionList = v1alpha1.ClusterServiceVersionList{}
	for _, _namespace := range namespaces {
		n_ClusterServiceVersionList := v1alpha1.ClusterServiceVersionList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/operators.coreos.com/clusterserviceversions/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/operators.coreos.com/clusterserviceversions/" + f.Name()
			_file, _ := ioutil.ReadFile(smcpYamlPath)
			_ClusterServiceVersion := v1alpha1.ClusterServiceVersion{}
			if err := yaml.Unmarshal([]byte(_file), &_ClusterServiceVersion); err != nil {
				fmt.Println("Error when trying to unmarshall file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_ClusterServiceVersionList.Items = append(n_ClusterServiceVersionList.Items, _ClusterServiceVersion)
		}
		for _, ClusterServiceVersion := range n_ClusterServiceVersionList.Items {
			if resourceName != "" && resourceName != ClusterServiceVersion.Name {
				continue
			}

			if outputFlag == "yaml" {
				n_ClusterServiceVersionList.Items = append(n_ClusterServiceVersionList.Items, ClusterServiceVersion)
				ClusterServiceVersionList.Items = append(ClusterServiceVersionList.Items, ClusterServiceVersion)
				continue
			}

			if outputFlag == "json" {
				n_ClusterServiceVersionList.Items = append(n_ClusterServiceVersionList.Items, ClusterServiceVersion)
				ClusterServiceVersionList.Items = append(ClusterServiceVersionList.Items, ClusterServiceVersion)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_ClusterServiceVersionList.Items = append(n_ClusterServiceVersionList.Items, ClusterServiceVersion)
				ClusterServiceVersionList.Items = append(ClusterServiceVersionList.Items, ClusterServiceVersion)
				continue
			}

			//name
			ClusterServiceVersionName := ClusterServiceVersion.Name
			//display
			display := ClusterServiceVersion.Spec.DisplayName
			//version
			version := ClusterServiceVersion.Spec.Version.Version.String()
			//replaces
			replaces := ClusterServiceVersion.Spec.Replaces
			//phase
			phase := string(ClusterServiceVersion.Status.Phase)

			labels := helpers.ExtractLabels(ClusterServiceVersion.GetLabels())
			_list := []string{ClusterServiceVersion.Namespace, ClusterServiceVersionName, display, version, replaces, phase}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 6, _list)

			if resourceName != "" && resourceName == ClusterServiceVersionName {
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

	if len(ClusterServiceVersionList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = ClusterServiceVersionList.Items[0]
	} else {
		resource = ClusterServiceVersionList
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

var ClusterServiceVersion = &cobra.Command{
	Use:     "clusterserviceversion",
	Aliases: []string{"csv", "clusterserviceversions", "clusterserviceversion.operators.coreos.com"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetClusterServiceVersion(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

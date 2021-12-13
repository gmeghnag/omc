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
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	v1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func GetInstallPlan(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "csv", "approval", "approved"}
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
	var InstallPlanList = v1alpha1.InstallPlanList{}
	for _, _namespace := range namespaces {
		n_InstallPlanList := v1alpha1.InstallPlanList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/operators.coreos.com/installplans/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/operators.coreos.com/installplans/" + f.Name()
			_file, err := ioutil.ReadFile(smcpYamlPath)
			if err != nil {
				fmt.Println(err.Error())
			}
			_InstallPlan := v1alpha1.InstallPlan{}
			if err := yaml.Unmarshal([]byte(_file), &_InstallPlan); err != nil {
				fmt.Println("Error when trying to unmarshal file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_InstallPlanList.Items = append(n_InstallPlanList.Items, _InstallPlan)
		}
		for _, InstallPlan := range n_InstallPlanList.Items {
			if resourceName != "" && resourceName != InstallPlan.Name {
				continue
			}

			if outputFlag == "yaml" {
				n_InstallPlanList.Items = append(n_InstallPlanList.Items, InstallPlan)
				InstallPlanList.Items = append(InstallPlanList.Items, InstallPlan)
				continue
			}

			if outputFlag == "json" {
				n_InstallPlanList.Items = append(n_InstallPlanList.Items, InstallPlan)
				InstallPlanList.Items = append(InstallPlanList.Items, InstallPlan)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_InstallPlanList.Items = append(n_InstallPlanList.Items, InstallPlan)
				InstallPlanList.Items = append(InstallPlanList.Items, InstallPlan)
				continue
			}

			//name
			InstallPlanName := InstallPlan.Name
			//package
			csv := ""
			if len(InstallPlan.Spec.ClusterServiceVersionNames) == 1 {
				csv = InstallPlan.Spec.ClusterServiceVersionNames[0]
			}
			if len(InstallPlan.Spec.ClusterServiceVersionNames) > 1 {
				csv = "[" + strings.Join(InstallPlan.Spec.ClusterServiceVersionNames, ", ") + "]"
			}
			//source
			approval := string(InstallPlan.Spec.Approval)
			//channel
			approved := strconv.FormatBool(InstallPlan.Spec.Approved)

			labels := helpers.ExtractLabels(InstallPlan.GetLabels())
			_list := []string{_namespace, InstallPlanName, csv, approval, approved}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == InstallPlanName {
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

	if len(InstallPlanList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = InstallPlanList.Items[0]
	} else {
		resource = InstallPlanList
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

var InstallPlan = &cobra.Command{
	Use:     "installplan",
	Aliases: []string{"ip", "installplans", "installplan.operators.coreos.com"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetInstallPlan(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

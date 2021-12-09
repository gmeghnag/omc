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

func GetSubscription(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "package", "source", "channel"}
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
	var SubscriptionList = v1alpha1.SubscriptionList{}
	for _, _namespace := range namespaces {
		n_SubscriptionList := v1alpha1.SubscriptionList{}
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_smcps, _ := ioutil.ReadDir(CurrentNamespacePath + "/operators.coreos.com/subscriptions/")
		for _, f := range _smcps {
			smcpYamlPath := CurrentNamespacePath + "/operators.coreos.com/subscriptions/" + f.Name()
			_file, err := ioutil.ReadFile(smcpYamlPath)
			if err != nil {
				fmt.Println(err.Error())
			}
			_Subscription := v1alpha1.Subscription{}
			if err := yaml.Unmarshal([]byte(_file), &_Subscription); err != nil {
				fmt.Println("Error when trying to unmarshal file: " + smcpYamlPath)
				os.Exit(1)
			}
			n_SubscriptionList.Items = append(n_SubscriptionList.Items, _Subscription)
		}
		for _, Subscription := range n_SubscriptionList.Items {
			if resourceName != "" && resourceName != Subscription.Name {
				continue
			}

			if outputFlag == "yaml" {
				n_SubscriptionList.Items = append(n_SubscriptionList.Items, Subscription)
				SubscriptionList.Items = append(SubscriptionList.Items, Subscription)
				continue
			}

			if outputFlag == "json" {
				n_SubscriptionList.Items = append(n_SubscriptionList.Items, Subscription)
				SubscriptionList.Items = append(SubscriptionList.Items, Subscription)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				n_SubscriptionList.Items = append(n_SubscriptionList.Items, Subscription)
				SubscriptionList.Items = append(SubscriptionList.Items, Subscription)
				continue
			}

			//name
			SubscriptionName := Subscription.Name
			//package
			subPackage := Subscription.Spec.Package
			//source
			source := Subscription.Spec.CatalogSource
			//channel
			channel := Subscription.Spec.Channel

			labels := helpers.ExtractLabels(Subscription.GetLabels())
			_list := []string{_namespace, SubscriptionName, subPackage, source, channel}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == SubscriptionName {
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

	if len(SubscriptionList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}
	var resource interface{}
	if resourceName != "" {
		resource = SubscriptionList.Items[0]
	} else {
		resource = SubscriptionList
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

var Subscription = &cobra.Command{
	Use:     "subscription",
	Aliases: []string{"sub", "subscriptions", "subscription.operators.coreos.com"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetSubscription(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

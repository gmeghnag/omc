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
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
)

type IngressesItems struct {
	ApiVersion string                  `json:"apiVersion"`
	Items      []*networkingv1.Ingress `json:"items"`
}

func GetIngresses(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "class", "hosts", "address", "port", "age"}

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
	var _IngressesList = IngressesItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items IngressesItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/networking.k8s.io/ingresses.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Fprintln(os.Stderr, "No resources found in "+_namespace+" namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+CurrentNamespacePath+"/networking.k8s.io/ingresses.yaml")
			os.Exit(1)
		}

		for _, Ingress := range _Items.Items {
			labels := helpers.ExtractLabels(Ingress.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}
			if resourceName != "" && resourceName != Ingress.Name {
				continue
			}

			if outputFlag == "name" {
				_IngressesList.Items = append(_IngressesList.Items, Ingress)
				fmt.Println("ingress.networking.k8s.io/" + Ingress.Name)
				continue
			}
			if outputFlag == "yaml" {
				_IngressesList.Items = append(_IngressesList.Items, Ingress)
				continue
			}

			if outputFlag == "json" {
				_IngressesList.Items = append(_IngressesList.Items, Ingress)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_IngressesList.Items = append(_IngressesList.Items, Ingress)
				continue
			}

			//name
			IngressName := Ingress.Name
			//class
			class := ""
			className := Ingress.Spec.IngressClassName
			if className == nil {
				class = "<none>"
			} else {
				class = *className
			}
			//host
			ingressRules := Ingress.Spec.Rules
			ingressHosts := ""
			for _, rule := range ingressRules {
				if string(rule.Host) != "" {
					ingressHosts += string(rule.Host) + ","
				}
			}
			if ingressHosts == "" {
				ingressHosts = "*"
			} else {
				ingressHosts = strings.TrimRight(ingressHosts, ",")
			}

			//address
			Ingressaddresses := ""
			for _, ingress := range Ingress.Status.LoadBalancer.Ingress {
				if string(ingress.Hostname) != "" {
					Ingressaddresses += string(ingress.Hostname) + ","
				}
			}
			if Ingressaddresses != "" {
				Ingressaddresses = strings.TrimRight(Ingressaddresses, ",")
			}
			IngressPorts := ""
			for _, rule := range ingressRules {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil {
						IngressPorts += strconv.Itoa(int(path.Backend.Service.Port.Number)) + ","
					}
				}
			}
			if IngressPorts != "" {
				IngressPorts = strings.TrimRight(IngressPorts, ",")
			}
			//age
			age := helpers.GetAge(CurrentNamespacePath+"/networking.k8s.io/ingresses.yaml", Ingress.GetCreationTimestamp())

			//labels
			_list := []string{Ingress.Namespace, IngressName, class, ingressHosts, Ingressaddresses, IngressPorts, age}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 7, _list)

			if resourceName != "" && resourceName == IngressName {
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

	if len(_IngressesList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _IngressesList.Items[0]
	} else {
		resource = _IngressesList
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

var Ingress = &cobra.Command{
	Use:     "ingress",
	Aliases: []string{"ingresses", "ing", "ingress.networking.k8s.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetIngresses(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

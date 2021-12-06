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
	"sort"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type NodesItems struct {
	ApiVersion string        `json:"apiVersion"`
	Items      []corev1.Node `json:"items"`
}

func getNodes(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	nodesFolderPath := currentContextPath + "/cluster-scoped-resources/core/nodes/"
	_nodes, _ := ioutil.ReadDir(nodesFolderPath)

	_headers := []string{"name", "status", "roles", "age", "version", "internal-ip", "external-ip", "os-image", "kernel-version", "container-runtime"}
	var data [][]string

	_NodesList := NodesItems{ApiVersion: "v1"}
	for _, f := range _nodes {
		nodeYamlPath := nodesFolderPath + f.Name()
		_file := helpers.ReadYaml(nodeYamlPath)
		Node := corev1.Node{}
		if err := yaml.Unmarshal([]byte(_file), &Node); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + nodeYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != Node.Name {
			continue
		}

		if outputFlag == "yaml" {
			_NodesList.Items = append(_NodesList.Items, Node)
			continue
		}

		if outputFlag == "json" {
			_NodesList.Items = append(_NodesList.Items, Node)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_NodesList.Items = append(_NodesList.Items, Node)
			continue
		}
		// STATUS
		NodeStatus := "NotReady"
		for _, condition := range Node.Status.Conditions {
			if condition.Type == "Ready" {
				if condition.Status == "True" {
					NodeStatus = "Ready"
				}
				break
			}
		}
		if Node.Spec.Unschedulable {
			NodeStatus += ",SchedulingDisabled"
		}

		//ROLE
		var NodeRoles []string
		NodeRole := ""
		for i := range Node.ObjectMeta.Labels {
			if strings.HasPrefix(i, "node-role.kubernetes.io/") {
				s := strings.Split(i, "/")
				NodeRoles = append(NodeRoles, s[1])
			}
		}
		sort.Strings(NodeRoles)
		if len(NodeRoles) == 0 {
			NodeRole = ""
		} else {
			NodeRole = strings.Join(NodeRoles, ",")
		}

		//AGE
		age := helpers.GetAge(nodeYamlPath, Node.GetCreationTimestamp())
		//ADDRESSES
		internalAddress := "<none>"
		externalAddress := "<none>"
		addresses := Node.Status.Addresses

		for _, add := range addresses {
			if add.Type == "InternalIP" {
				internalAddress = add.Address
			}
			if add.Type == "ExternalIP" {
				externalAddress = add.Address
			}
		}
		labels := helpers.ExtractLabels(Node.GetLabels())
		_list := []string{Node.Name, NodeStatus, NodeRole, age, Node.Status.NodeInfo.KubeletVersion, internalAddress, externalAddress, Node.Status.NodeInfo.OSImage, Node.Status.NodeInfo.KernelVersion, Node.Status.NodeInfo.ContainerRuntimeVersion}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 5, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:5] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	var resource interface{}
	if resourceName != "" {
		resource = _NodesList.Items[0]
	} else {
		resource = _NodesList
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

var Node = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getNodes(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

/*
Copyright © 2021 Bram Verschueren <bverschueren@redhat.com>

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
package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	networkv1 "github.com/openshift/api/network/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

func getClusterNetwork(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	clusterNetworksFolderPath := currentContextPath + "/cluster-scoped-resources/network.openshift.io/clusternetworks/"
	_clusterNetworks, _ := ioutil.ReadDir(clusterNetworksFolderPath)

	_headers := []string{"name", "cluster network", "service network", "plugin name"}
	var data [][]string

	_ClusterNetworkList := networkv1.ClusterNetworkList{}
	for _, f := range _clusterNetworks {
		clusterNetworkYamlPath := clusterNetworksFolderPath + f.Name()
		_file, _ := ioutil.ReadFile(clusterNetworkYamlPath)
		ClusterNetwork := networkv1.ClusterNetwork{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterNetwork); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + clusterNetworkYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(ClusterNetwork.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != ClusterNetwork.Name {
			continue
		}

		if outputFlag == "name" {
			_ClusterNetworkList.Items = append(_ClusterNetworkList.Items, ClusterNetwork)
			fmt.Println("clusternetwork.config.openshift.io/" + ClusterNetwork.Name)
			continue
		}

		if outputFlag == "yaml" {
			_ClusterNetworkList.Items = append(_ClusterNetworkList.Items, ClusterNetwork)
			continue
		}

		if outputFlag == "json" {
			_ClusterNetworkList.Items = append(_ClusterNetworkList.Items, ClusterNetwork)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ClusterNetworkList.Items = append(_ClusterNetworkList.Items, ClusterNetwork)
			continue
		}

		clusterNetworkName := ClusterNetwork.Name
		clusterNetworkCIDR := ClusterNetwork.Network
		serviceNetwork := ClusterNetwork.ServiceNetwork
		pluginName := ClusterNetwork.PluginName

		_list := []string{clusterNetworkName, clusterNetworkCIDR, serviceNetwork, pluginName}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 4, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:4] // -A
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
		resource = _ClusterNetworkList.Items[0]
	} else {
		resource = _ClusterNetworkList
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

var ClusterNetwork = &cobra.Command{
	Use:     "clusternetwork",
	Aliases: []string{"clusternetworks", "clusternetworks.network.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterNetwork(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

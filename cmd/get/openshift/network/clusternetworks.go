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

type clusterNetworksItems struct {
	ApiVersion string                     `json:"apiVersion"`
	Items      []networkv1.ClusterNetwork `json:"items"`
}

func getClusterNetwork(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	// There is only one clusternetwork per cluster, therefore the must gather
	// only contains a single clusternetwork rather than a list. Do not take
	// this file as an example because most of what you see is exceptional.

	if resourceName != "" && resourceName != "default" {
		fmt.Println("omc only supports the \"default\" clusternetwork. Try omc get clusternetwork or omc get clusternetwork default")
		os.Exit(1)
	}

	clusterNetworksYamlPath := currentContextPath + "/cluster-scoped-resources/network.openshift.io/clusternetworks/default.yaml"

	_file, _ := ioutil.ReadFile(clusterNetworksYamlPath)

	clusterNetwork := networkv1.ClusterNetwork{}
	if err := yaml.Unmarshal([]byte(_file), &clusterNetwork); err != nil {
		fmt.Println("Error when trying to unmarshal file: " + clusterNetworksYamlPath)
		os.Exit(1)
	}

	if outputFlag == "" || outputFlag == "wide" {
		headers := []string{"NAME", "CLUSTER NETWORK", "SERVICE NETWORK", "PLUGIN NAME"}
		clusterNetworkName := clusterNetwork.Name
		clusterNetworkCIDR := clusterNetwork.Network
		serviceNetwork := clusterNetwork.ServiceNetwork
		pluginName := clusterNetwork.PluginName
		labels := helpers.ExtractLabels(clusterNetwork.GetLabels())

		cn := []string{clusterNetworkName, clusterNetworkCIDR, serviceNetwork, pluginName}

		if showLabels {
			headers = append(headers, "labels")
			cn = append(cn, labels)
		}

		data := [][]string{cn}
		helpers.PrintTable(headers, data)
	}

	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(clusterNetwork)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(clusterNetwork, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(clusterNetwork, jsonPathTemplate)
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

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
package ovn

import (
	"fmt"
	"os"
	"slices"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"strings"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

var SubnetsCmd = &cobra.Command{
	Use:     "subnets",
	Aliases: []string{"subnet"},
	Short:   "Retrieve the ovn nodes and subnets they are providing.",
	Run: func(cmd *cobra.Command, args []string) {

		nodesFolderPath := vars.MustGatherRootPath + "/cluster-scoped-resources/core/nodes/"
		_nodes, _ := os.ReadDir(nodesFolderPath)

		var data [][]string
		var nodeSubnetInHeaders, nodeGatewayRouterIpInHeaders, nodeTransitSwitchIpInHeaders, nodeMasqueradeSubnetInHeaders bool
		headers := []string{"HOST/NODE", "ROLE"}
		for _, f := range _nodes {
			nodeYamlPath := nodesFolderPath + f.Name()
			_file := helpers.ReadYaml(nodeYamlPath)
			Node := corev1.Node{}
			if err := yaml.Unmarshal([]byte(_file), &Node); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+nodeYamlPath)
				os.Exit(1)
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
			slices.Sort(NodeRoles)
			NodeRole = strings.Join(NodeRoles, ",")

			row := []string{Node.Name, NodeRole}

			nodeSubnet := ""
			nodeSubnetStrMap := Node.ObjectMeta.Annotations["k8s.ovn.org/node-subnets"]
			if nodeSubnetStrMap != "" {
				var subnet map[string]string
				if err := yaml.Unmarshal([]byte(nodeSubnetStrMap), &subnet); err != nil {
					var subnet map[string][]string
					if err := yaml.Unmarshal([]byte(nodeSubnetStrMap), &subnet); err != nil {
						panic(err)
					}
					nodeSubnet = strings.Join(subnet["default"], ",")
				} else {
					nodeSubnet = subnet["default"]
				}
			}

			if nodeSubnet != "" {
				row = append(row, nodeSubnet)
				if !nodeSubnetInHeaders {
					headers = append(headers, "NODE SUBNET")
					nodeSubnetInHeaders = true
				}
			}

			nodeGatewayRouterIp := ""
			nodeGatewayRouterIpStrMap := Node.ObjectMeta.Annotations["k8s.ovn.org/node-gateway-router-lrp-ifaddr"]
			if nodeGatewayRouterIpStrMap != "" {
				var gatewayRouterIp map[string]string
				if err := yaml.Unmarshal([]byte(nodeGatewayRouterIpStrMap), &gatewayRouterIp); err != nil {
					panic(err)
				}
				nodeGatewayRouterIp = gatewayRouterIp["ipv4"]
			}

			nodeGatewayRouterIpStrMapMap := Node.ObjectMeta.Annotations["k8s.ovn.org/node-gateway-router-lrp-ifaddrs"]
			if nodeGatewayRouterIpStrMapMap != "" {
				var gatewayRouterIps map[string]map[string]string
				if err := yaml.Unmarshal([]byte(nodeGatewayRouterIpStrMapMap), &gatewayRouterIps); err != nil {
					panic(err)
				}
				if gatewayRouterIp, ok := gatewayRouterIps["default"]; ok {
					nodeGatewayRouterIp = gatewayRouterIp["ipv4"]
				}
			}

			if nodeGatewayRouterIp != "" {
				row = append(row, nodeGatewayRouterIp)
				if !nodeGatewayRouterIpInHeaders {
					headers = append(headers, "NODE GW-ROUTER-IP")
					nodeGatewayRouterIpInHeaders = true
				}
			}

			nodeTransitSwitchIp := ""
			nodeTransitSwitchIpStrMap := Node.ObjectMeta.Annotations["k8s.ovn.org/node-transit-switch-port-ifaddr"]
			if nodeTransitSwitchIpStrMap != "" {
				var transitSwitchIp map[string]string
				if err := yaml.Unmarshal([]byte(nodeTransitSwitchIpStrMap), &transitSwitchIp); err != nil {
					panic(err)
				}
				nodeTransitSwitchIp = transitSwitchIp["ipv4"]
			}

			if nodeTransitSwitchIp != "" {
				row = append(row, nodeTransitSwitchIp)
				if !nodeTransitSwitchIpInHeaders {
					headers = append(headers, "NODE TRANSIT-SWITCH-IP")
					nodeTransitSwitchIpInHeaders = true
				}
			}

			if vars.OutputStringVar == "wide" {

				nodeMasqueradeSubnet := ""
				nodeMasqueradeSubnetStrMap := Node.ObjectMeta.Annotations["k8s.ovn.org/node-masquerade-subnet"]
				if nodeMasqueradeSubnetStrMap != "" {
					var masqueradeSubnet map[string]string
					if err := yaml.Unmarshal([]byte(nodeMasqueradeSubnetStrMap), &masqueradeSubnet); err != nil {
						panic(err)
					}
					nodeMasqueradeSubnet = masqueradeSubnet["ipv4"]
				}

				if nodeMasqueradeSubnet != "" {
					row = append(row, nodeMasqueradeSubnet)
					if !nodeMasqueradeSubnetInHeaders {
						headers = append(headers, "NODE MASQUERADE-SUBNET")
						nodeMasqueradeSubnetInHeaders = true
					}
				}

			}

			data = append(data, row)

		}
		helpers.PrintTable(headers, data)
	},
}

func init() {
	SubnetsCmd.Flags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format: wide.")
}

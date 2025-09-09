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

var HostnetinfoCmd = &cobra.Command{
	Use:     "hostnetinfo",
	Aliases: []string{"hostnetinfo", "hostnetwork"},
	Short:   "Retrieve the host network information that OVN-Kubernetes needs to store about its nodes.",
	Run: func(cmd *cobra.Command, args []string) {

		nodesFolderPath := vars.MustGatherRootPath + "/cluster-scoped-resources/core/nodes/"
		_nodes, _ := os.ReadDir(nodesFolderPath)

		var data [][]string
		var ipv4InHeaders, ipv6InHeaders, gatewayIPInHeaders, primaryIfAddrInHeaders bool
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
			ipv4String := Node.ObjectMeta.Annotations["alpha.kubernetes.io/provided-node-ip"]
			ipv6String := ""
			var ipsArray []string
			var ipv4Array []string
			var ipv6Array []string
			hostAddresses := Node.ObjectMeta.Annotations["k8s.ovn.org/host-addresses"]
			if hostAddresses != "" {
				err := yaml.Unmarshal([]byte(hostAddresses), &ipsArray)
				if err != nil {
					panic(err)
				}
				for _, ip := range ipsArray {
					if strings.Contains(ip, ":") {
						ipv6Array = append(ipv6Array, ip)
					} else {
						ipv4Array = append(ipv4Array, ip)
					}
				}
			}
			// "k8s.ovn.org/host-addresses" was renamed to "k8s.ovn.org/host-cidrs" in 4.14
			hostCIDRS := Node.ObjectMeta.Annotations["k8s.ovn.org/host-cidrs"]
			if hostCIDRS != "" {
				err := yaml.Unmarshal([]byte(hostCIDRS), &ipsArray)
				if err != nil {
					panic(err)
				}
				for _, ip := range ipsArray {
					if strings.Contains(ip, ":") {
						ipv6Array = append(ipv6Array, ip)
					} else {
						ipv4Array = append(ipv4Array, ip)
					}
				}
			}
			if len(ipv4Array) != 0 {
				ipv4String = strings.Join(ipv4Array, ",")
			}
			if len(ipv6Array) != 0 {
				ipv6String = strings.Join(ipv6Array, ",")
			}
			if ipv6String != "" {
				row = append(row, ipv6String)
				if !ipv6InHeaders {
					headers = append(headers, "HOST IPV6-ADDRESSES")
					ipv6InHeaders = true
				}
			}
			if ipv4String != "" {
				row = append(row, ipv4String)
				if !ipv4InHeaders {
					headers = append(headers, "HOST IP-ADDRESSES")
					ipv4InHeaders = true
				}
			}

			primaryIfAddr := ""
			primaryIfAddrStrMap := Node.ObjectMeta.Annotations["k8s.ovn.org/node-primary-ifaddr"]
			if primaryIfAddrStrMap != "" {
				var ifaddr map[string]string
				if err := yaml.Unmarshal([]byte(primaryIfAddrStrMap), &ifaddr); err != nil {
					panic(err)
				}
				primaryIfAddr = ifaddr["ipv4"]
			}
			if primaryIfAddr != "" {
				row = append(row, primaryIfAddr)
				if !primaryIfAddrInHeaders {
					headers = append(headers, "PRIMARY IF-ADDRESS")
					primaryIfAddrInHeaders = true
				}
			}

			gatewayIP := ""
			GatewayConfigString := Node.ObjectMeta.Annotations["k8s.ovn.org/l3-gateway-config"]
			if GatewayConfigString != "" {
				var GatewayConf GatewayConfig
				if err := yaml.Unmarshal([]byte(GatewayConfigString), &GatewayConf); err != nil {
					panic(err)
				}
				gatewayIP = strings.Join(GatewayConf.Default.NextHops, ",")
			}
			if gatewayIP != "" {
				row = append(row, gatewayIP)
				if !gatewayIPInHeaders {
					headers = append(headers, "HOST GATEWAY-IP")
					gatewayIPInHeaders = true
				}
			}

			// if vars.OutputStringVar == "wide" {
			// If something is to be displayed only with "wide" output, include here.
			// }
			data = append(data, row)

		}
		helpers.PrintTable(headers, data)
	},
}

type GatewayConfig struct {
	Default Config `json:"default"`
}

type Config struct {
	Mode           string   `json:"mode"`
	InterfaceID    string   `json:"interface-id"`
	MacAddress     string   `json:"mac-address"`
	MacAddresses   []string `json:"ip-addresses"`
	IpAddresses    string   `json:"ip-address"`
	NextHops       []string `json:"next-hops"`
	NextHop        string   `json:"next-hop"`
	NodePortEnable string   `json:"node-port-enable"`
	VlanId         string   `json:"vlan-id"`
}

func init() {
	HostnetinfoCmd.Flags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format: wide.")
}

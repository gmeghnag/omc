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

var NodeExtraInfoCmd = &cobra.Command{
	Use:     "extrainfo",
	Aliases: []string{"node-extrainfo"},
	Short:   "Retrieve some extra information OVN-Kubernetes needs to store about nodes.",
	Run: func(cmd *cobra.Command, args []string) {

		nodesFolderPath := vars.MustGatherRootPath + "/cluster-scoped-resources/core/nodes/"
		_nodes, _ := os.ReadDir(nodesFolderPath)

		var data [][]string
		var nodeIdInHeaders, nodeChassisIdInHeaders, nodeZoneNameInHeaders, nodeRemoteZoneMigratedInHeaders bool
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

			nodeId := Node.ObjectMeta.Annotations["k8s.ovn.org/node-id"]
			if nodeId != "" {
				row = append(row, nodeId)
				if !nodeIdInHeaders {
					headers = append(headers, "NODE ID")
					nodeIdInHeaders = true
				}
			}

			nodeChassisId := Node.ObjectMeta.Annotations["k8s.ovn.org/node-chassis-id"]
			if nodeChassisId != "" {
				row = append(row, nodeChassisId)
				if !nodeChassisIdInHeaders {
					headers = append(headers, "NODE CHASSIS-ID")
					nodeChassisIdInHeaders = true
				}
			}

			if vars.OutputStringVar == "wide" {

				nodeZoneName := Node.ObjectMeta.Annotations["k8s.ovn.org/zone-name"]
				if nodeZoneName != "" {
					row = append(row, nodeZoneName)
					if !nodeZoneNameInHeaders {
						headers = append(headers, "NODE ZONE-NAME")
						nodeZoneNameInHeaders = true
					}
				}

				nodeRemoteZoneMigrated := Node.ObjectMeta.Annotations["k8s.ovn.org/remote-zone-migrated"]
				if nodeRemoteZoneMigrated != "" {
					row = append(row, nodeRemoteZoneMigrated)
					if !nodeRemoteZoneMigratedInHeaders {
						headers = append(headers, "NODE REMOTE-ZONE-MIGRATED")
						nodeRemoteZoneMigratedInHeaders = true
					}
				}
			}
			data = append(data, row)

		}
		helpers.PrintTable(headers, data)
	},
}

func init() {
	NodeExtraInfoCmd.Flags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format: wide.")
}

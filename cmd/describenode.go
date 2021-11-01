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
package cmd

import (
	"fmt"
	"omc/cmd/helpers"
	"omc/types"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	describe "k8s.io/kubectl/pkg/describe"
	"sigs.k8s.io/yaml"
)

func describeNodeCommand(currentContextPath string, defaultConfigNamespace string, resourceName string) {
	nodePath := currentContextPath + "/cluster-scoped-resources/core/nodes/" + resourceName + ".yaml"
	Node := corev1.Node{}
	_file := helpers.ReadYaml(nodePath)
	if err := yaml.Unmarshal([]byte(_file), &Node); err != nil {
		fmt.Println("Error when trying to unmarshall file " + nodePath)
		os.Exit(1)
	}
	fake := fake.NewSimpleClientset(&Node)
	c := &types.DescribeClient{Namespace: defaultConfigNamespace, Interface: fake}
	d := describe.NodeDescriber{c}
	out, _ := d.Describe(defaultConfigNamespace, resourceName, describe.DescriberSettings{ShowEvents: true})
	fmt.Printf(out)
}

// describeCmd represents the etcd command
var describeNodeCmd = &cobra.Command{
	Use:     "node",
	Short:   "describe command",
	Aliases: []string{"nodes"},
	Run: func(cmd *cobra.Command, args []string) {
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		if namespaceFlag != "" {
			defaultConfigNamespace, _ = rootCmd.PersistentFlags().GetString("namespace")
		}
		if len(args) == 0 || len(args) > 1 {
			fmt.Println("Expected one arguments, found: " + strconv.Itoa(len(args)) + ".")
			os.Exit(1)
		}
		resourceName := strings.ToLower(args[0])
		describeNodeCommand(currentContextPath, defaultConfigNamespace, resourceName)

	},
}

func init() {
	describeCmd.AddCommand(describeNodeCmd)
}

//func (templater *templater) HelpFunc() func(*cobra.Command, []string) {
//	return func(c *cobra.Command, s []string) {
//		t := template.New("help")
//		t.Funcs(templater.templateFuncs())
//		template.Must(t.Parse(templater.HelpTemplate))
//		out := term.NewResponsiveWriter(c.OutOrStdout())
//		err := t.Execute(out, c)
//		if err != nil {
//			c.Println(err)
//		}
//	}
//}

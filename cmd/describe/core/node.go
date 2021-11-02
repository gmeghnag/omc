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
	"fmt"
	"omc/cmd/helpers"
	"omc/types"
	"omc/vars"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	desc "k8s.io/kubectl/pkg/describe"
	"sigs.k8s.io/yaml"
)

func describeNode(currentContextPath string, namespace string, resourceName string) {
	nodePath := currentContextPath + "/cluster-scoped-resources/core/nodes/" + resourceName + ".yaml"
	Node := corev1.Node{}
	_file := helpers.ReadYaml(nodePath)
	if err := yaml.Unmarshal([]byte(_file), &Node); err != nil {
		fmt.Println("Error when trying to unmarshall file " + nodePath)
		os.Exit(1)
	}
	fake := fake.NewSimpleClientset(&Node)
	c := &types.DescribeClient{Namespace: namespace, Interface: fake}
	d := desc.NodeDescriber{c}
	out, _ := d.Describe(namespace, resourceName, desc.DescriberSettings{ShowEvents: false})
	fmt.Printf(out)
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
		describeNode(vars.MustGatherRootPath, vars.Namespace, resourceName)
	},
}

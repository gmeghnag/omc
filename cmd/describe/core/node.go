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
	"io/ioutil"
	"os"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	desc "k8s.io/kubectl/pkg/describe"
	"sigs.k8s.io/yaml"
)

func describeNode(currentContextPath string, namespace string, args []string) {
	resourceDir := currentContextPath + "/cluster-scoped-resources/core/nodes"
	resourcesFiles, _ := ioutil.ReadDir(resourceDir)
	for _, f := range resourcesFiles {
		resourceYamlPath := resourceDir + "/" + f.Name()
		_file, _ := ioutil.ReadFile(resourceYamlPath)
		_Node := corev1.Node{}
		if err := yaml.Unmarshal(_file, &_Node); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
			os.Exit(1)
		}
		if len(args) > 0 && helpers.StringInSlice(_Node.GetName(), args) {
			fake := fake.NewSimpleClientset(&_Node)
			c := &types.DescribeClient{Namespace: namespace, Interface: fake}
			d := desc.NodeDescriber{c}
			out, _ := d.Describe(namespace, _Node.GetName(), desc.DescriberSettings{ShowEvents: false})
			fmt.Printf("%s", out)
		}
	}
}

var Node = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		describeNode(vars.MustGatherRootPath, vars.Namespace, args)
	},
}

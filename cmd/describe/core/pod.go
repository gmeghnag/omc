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
	"os"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/describe"
	"sigs.k8s.io/yaml"
)

func describePod(currentContextPath string, defaultConfigNamespace string, args []string) {
	podResources := currentContextPath + "/namespaces/" + defaultConfigNamespace + "/core/pods.yaml"
	_file, err := os.ReadFile(podResources)
	if err == nil {
		PodList := corev1.PodList{}
		if err := yaml.Unmarshal(_file, &PodList); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+podResources)
			os.Exit(1)
		}
		for _, pod := range PodList.Items {
			if len(args) == 0 || (len(args) > 0 && helpers.StringInSlice(pod.GetName(), args)) {
				fake := fake.NewSimpleClientset(&pod)
				c := &types.DescribeClient{Namespace: defaultConfigNamespace, Interface: fake}
				d := describe.PodDescriber{c}
				out, _ := d.Describe(defaultConfigNamespace, pod.GetName(), describe.DescriberSettings{ShowEvents: false})
				fmt.Printf("%s", out)
			}
		}
	} else {
		podsDir := fmt.Sprintf("%s/namespaces/%s/pods", vars.MustGatherRootPath, defaultConfigNamespace)
		pods, rErr := os.ReadDir(podsDir)
		if rErr != nil {
			klog.V(3).ErrorS(err, "Failed to read resources:")
		}
		for _, pod := range pods {
			podName := pod.Name()
			podPath := fmt.Sprintf("%s/%s/%s.yaml", podsDir, podName, podName)
			_file, err := os.ReadFile(podPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: %s\n", podPath, err)
				os.Exit(1)
			}
			var podItem corev1.Pod
			if err := yaml.Unmarshal(_file, &podItem); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+podPath)
				os.Exit(1)
			}
			if len(args) == 0 || (len(args) > 0 && helpers.StringInSlice(podItem.GetName(), args)) {
				fake := fake.NewSimpleClientset(&podItem)
				c := &types.DescribeClient{Namespace: defaultConfigNamespace, Interface: fake}
				d := describe.PodDescriber{c}
				out, _ := d.Describe(defaultConfigNamespace, podItem.GetName(), describe.DescriberSettings{ShowEvents: false})
				fmt.Printf("%s", out)
			}
		}
	}
}

var Pod = &cobra.Command{
	Use:     "pod",
	Aliases: []string{"po", "pods"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		describePod(vars.MustGatherRootPath, vars.Namespace, args)
	},
}

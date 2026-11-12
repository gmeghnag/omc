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
	"github.com/gmeghnag/omc/pkg/mustgather"
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
	namespacePath := currentContextPath + "/namespaces/" + defaultConfigNamespace
	// helpers.PodObjects resolves every pod storage layout a must-gather may
	// use (aggregated core/pods.yaml, nested pods/<pod>/<pod>.yaml, flat
	// core/pods/<pod>.yaml), so describe shows the same pods get does.
	objects, err := helpers.PodObjects(namespacePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(objects) == 0 {
		klog.V(3).Infof("no pods found under %s", namespacePath)
	}
	for _, object := range objects {
		var pod corev1.Pod
		if err := yaml.Unmarshal(object.Data, &pod); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+object.Path)
			os.Exit(1)
		}
		if len(args) > 0 && !helpers.StringInSlice(pod.GetName(), args) {
			continue
		}
		fake := fake.NewSimpleClientset(&pod)
		c := &types.DescribeClient{Namespace: defaultConfigNamespace, Interface: fake}
		d := describe.PodDescriber{c}
		out, _ := d.Describe(defaultConfigNamespace, pod.GetName(), describe.DescriberSettings{ShowEvents: false})
		fmt.Printf("%s", out)
	}
}

var Pod = &cobra.Command{
	Use:     "pod",
	Aliases: []string{"po", "pods"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		describePodMulti(vars.Namespace, args)
	},
}

// describePodMulti resolves each requested pod to the most recent must-gather
// that contains it, so describe shows the same capture get and logs do. For a
// single-must-gather context, or when describing every pod (no names given), it
// reads the most recent root directly.
func describePodMulti(namespace string, args []string) {
	roots := vars.MustGatherRootPaths
	if len(roots) <= 1 {
		describePod(vars.MustGatherRootPath, namespace, args)
		return
	}
	if len(args) == 0 {
		describePod(roots[0], namespace, args)
		return
	}
	for _, name := range args {
		root, _ := mustgather.ResolveRootBy(roots, func(root string) bool {
			return podExistsInRoot(root, namespace, name)
		})
		describePod(root, namespace, []string{name})
	}
}

// podExistsInRoot reports whether a must-gather root contains the named pod, in
// any of the per-pod or aggregated layouts helpers.PodExists knows about, so
// describe resolves to the same capture get would show regardless of layout.
func podExistsInRoot(root, namespace, name string) bool {
	return helpers.PodExists(root+"/namespaces/"+namespace, name)
}

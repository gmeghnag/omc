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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"strconv"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type ConfigMapsItems struct {
	ApiVersion string                       `json:"apiVersion"`
	Items      []*unstructured.Unstructured `json:"items"`
}

func GetConfigMaps(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, out *[]*unstructured.Unstructured) {
	var namespaces []string
	if allNamespacesFlag == true {
		namespace = "all"
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, namespace)
	}

	for _, _namespace := range namespaces {
		var _Items ConfigMapsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/configmaps.yaml")
		if err != nil && !allNamespacesFlag {
			continue
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+CurrentNamespacePath+"/core/configmaps.yaml")
			os.Exit(1)
		}
		for _, ConfigMap := range _Items.Items {
			labels := helpers.ExtractLabels(ConfigMap.GetLabels())
			if vars.LabelSelectorStringVar != "" {
				if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
					continue
				}
			}
			if resourceName != "" && resourceName != ConfigMap.GetName() {
				continue
			}
			*out = append(*out, ConfigMap)
		}
	}
}

var ConfigMap = &cobra.Command{
	Use:     "configmap",
	Aliases: []string{"cm", "configmaps"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		var resources []*unstructured.Unstructured
		GetConfigMaps(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, &resources)
		if len(resources) == 0 {
			fmt.Fprintln(os.Stderr, "No resources found.")
			os.Exit(0)
		}
		_headers := []string{"namespace", "name", "data", "age"}
		allResources := false
		var data [][]string
		for _, ConfigMap := range resources {
			labels := helpers.ExtractLabels(ConfigMap.GetLabels())

			//name
			ConfigMapName := ConfigMap.GetName()
			if allResources {
				ConfigMapName = "configmap/" + ConfigMapName
			}
			//data
			var cm corev1.ConfigMap
			_ = runtime.DefaultUnstructuredConverter.FromUnstructured(ConfigMap.Object, &cm)
			configmapData := strconv.Itoa(len(cm.Data))

			//age
			age := helpers.GetAge(vars.MustGatherRootPath+"/namespaces/"+ConfigMap.GetNamespace()+"/core/", ConfigMap.GetCreationTimestamp())

			_list := []string{ConfigMap.GetNamespace(), ConfigMapName, configmapData, age}
			data = helpers.GetData(data, vars.AllNamespaceBoolVar, vars.ShowLabelsBoolVar, labels, vars.OutputStringVar, 4, _list)

			if resourceName != "" && resourceName == ConfigMapName {
				break
			}
		}
		// ugly hack to get single item out of the slice
		//  TODO: handle this is helpets.PrintOutput
		var resourceSliceOrSingle interface{}
		if resourceName == "" {
			// for backward-compability print this as a ConfigMapsItems
			resourceSliceOrSingle = ConfigMapsItems{ApiVersion: "v1", Items: resources}
		} else {
			resourceSliceOrSingle = resources[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		helpers.PrintOutput(resourceSliceOrSingle, 4, vars.OutputStringVar, resourceName, vars.AllNamespaceBoolVar, vars.ShowLabelsBoolVar, _headers, data, jsonPathTemplate)
	},
}

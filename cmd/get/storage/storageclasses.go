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
package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	storagev1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/yaml"
)

type getStorageClassesItems struct {
	ApiVersion string                   `json:"apiVersion"`
	Items      []storagev1.StorageClass `json:"items"`
}

func getStorageClasses(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	storageclassesFolderPath := currentContextPath + "/cluster-scoped-resources/storage.k8s.io/storageclasses/"
	_storageclasses, _ := ioutil.ReadDir(storageclassesFolderPath)

	_headers := []string{"name", "provisioner", "reclaimpolicy", "volumebindingmode", "allowvolumeexpansion", "age"}
	var data [][]string

	_getStorageClassesList := getStorageClassesItems{ApiVersion: "v1"}
	for _, f := range _storageclasses {
		storageclassYamlPath := storageclassesFolderPath + f.Name()
		_file := helpers.ReadYaml(storageclassYamlPath)
		StorageClass := storagev1.StorageClass{}
		if err := yaml.Unmarshal([]byte(_file), &StorageClass); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+storageclassYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(StorageClass.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != StorageClass.Name {
			continue
		}

		if outputFlag == "name" {
			_getStorageClassesList.Items = append(_getStorageClassesList.Items, StorageClass)
			fmt.Println("storageclass.storage.k8s.io/" + StorageClass.Name)
			continue
		}

		if outputFlag == "yaml" {
			_getStorageClassesList.Items = append(_getStorageClassesList.Items, StorageClass)
			continue
		}

		if outputFlag == "json" {
			_getStorageClassesList.Items = append(_getStorageClassesList.Items, StorageClass)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_getStorageClassesList.Items = append(_getStorageClassesList.Items, StorageClass)
			continue
		}
		// NAME
		StorageClassName := StorageClass.Name
		for k, v := range StorageClass.GetAnnotations() {
			if k == "storageclass.kubernetes.io/is-default-class" && v == "true" {
				StorageClassName += " (default)"
				break
			}
		}
		// PROVISIONER
		provisioner := StorageClass.Provisioner
		// RECLAIMPOLICY
		reclaimPolicy := string(*StorageClass.ReclaimPolicy)
		//VOLUMEBINDINGMODE
		volumeBindingMode := string(*StorageClass.VolumeBindingMode)
		//ALLOWVOLUMEEXPANSION
		allowVolumeExpansion := "true"
		if StorageClass.AllowVolumeExpansion == nil {
			allowVolumeExpansion = "false"
		} else {
			allowVolumeExpansion = strconv.FormatBool(*StorageClass.AllowVolumeExpansion)
		}

		//AGE
		age := helpers.GetAge(storageclassYamlPath, StorageClass.GetCreationTimestamp())

		_list := []string{StorageClassName, provisioner, reclaimPolicy, volumeBindingMode, allowVolumeExpansion, age}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 6, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:6] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	var resource interface{}
	if resourceName != "" {
		resource = _getStorageClassesList.Items[0]
	} else {
		resource = _getStorageClassesList
	}
	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(resource)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(resource, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(resource, jsonPathTemplate)
	}
	return false
}

var StorageClass = &cobra.Command{
	Use:     "storageclass",
	Aliases: []string{"storageclasses", "sc", "storageclass.storage", "storageclass.storage.k8s.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getStorageClasses(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

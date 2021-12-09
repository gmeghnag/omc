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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type PersistentVolumeClaimsItems struct {
	ApiVersion string                          `json:"apiVersion"`
	Items      []*corev1.PersistentVolumeClaim `json:"items"`
}

func getPersistentVolumeClaims(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "status", "volume", "capacity", "access modes", "storageclass", "age", "volume mode"}
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

	var data [][]string
	var _PersistentVolumeClaimsList = PersistentVolumeClaimsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items PersistentVolumeClaimsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/persistentvolumeclaims.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshal file " + CurrentNamespacePath + "/core/persistentvolumeclaims.yaml")
			os.Exit(1)
		}

		for _, PersistentVolumeClaim := range _Items.Items {
			if resourceName != "" && resourceName != PersistentVolumeClaim.Name {
				continue
			}

			if outputFlag == "yaml" {
				_PersistentVolumeClaimsList.Items = append(_PersistentVolumeClaimsList.Items, PersistentVolumeClaim)
				continue
			}

			if outputFlag == "json" {
				_PersistentVolumeClaimsList.Items = append(_PersistentVolumeClaimsList.Items, PersistentVolumeClaim)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_PersistentVolumeClaimsList.Items = append(_PersistentVolumeClaimsList.Items, PersistentVolumeClaim)
				continue
			}

			//name
			name := PersistentVolumeClaim.Name

			//status
			status := string(PersistentVolumeClaim.Status.Phase)

			//volume
			volume := string(PersistentVolumeClaim.Spec.VolumeName)

			//capacity
			capacity := PersistentVolumeClaim.Status.Capacity.Storage().String()
			if capacity == "0" {
				capacity = ""
			}

			//access mode
			accessMode := ""
			for _, am := range PersistentVolumeClaim.Spec.AccessModes {
				if am == "ReadWriteOnce" {
					accessMode += "RWO,"
				}
				if am == "ReadWriteMany" {
					accessMode += "RWX,"
				}
				if am == "ReadOnlyMany" {
					accessMode += "ROX,"
				}
			}
			accessMode = strings.TrimSuffix(accessMode, ",")

			//storage class
			storageClass := ""
			if PersistentVolumeClaim.Spec.StorageClassName != nil {
				storageClass = *PersistentVolumeClaim.Spec.StorageClassName
			}

			//age
			age := helpers.GetAge(CurrentNamespacePath+"/core/persistentvolumeclaims.yaml", PersistentVolumeClaim.GetCreationTimestamp())
			//volumemode
			volumeMode := "Filesystem"
			if string(*PersistentVolumeClaim.Spec.VolumeMode) == "Block" {
				volumeMode = "Block"
			}

			//labels
			labels := helpers.ExtractLabels(PersistentVolumeClaim.GetLabels())
			_list := []string{PersistentVolumeClaim.Namespace, name, status, volume, capacity, accessMode, storageClass, age, volumeMode}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 8, _list)

			if resourceName != "" && resourceName == PersistentVolumeClaim.Name {
				break
			}
		}
		if namespace != "" && _namespace == namespace {
			break
		}
	}

	if (outputFlag == "" || outputFlag == "wide") && len(data) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var headers []string
	if outputFlag == "" {
		if allNamespacesFlag == true {
			headers = _headers[0:8]
		} else {
			headers = _headers[1:8]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		if allNamespacesFlag == true {
			headers = _headers
		} else {
			headers = _headers[1:]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}

	if len(_PersistentVolumeClaimsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _PersistentVolumeClaimsList.Items[0]
	} else {
		resource = _PersistentVolumeClaimsList
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

var PersistentVolumeClaim = &cobra.Command{
	Use:     "persistentvolumeclaim",
	Aliases: []string{"persistentvolumeclaims", "pvc"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getPersistentVolumeClaims(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

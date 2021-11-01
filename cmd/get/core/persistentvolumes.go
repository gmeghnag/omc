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
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type PersistentVolumesItems struct {
	ApiVersion string                    `json:"apiVersion"`
	Items      []corev1.PersistentVolume `json:"items"`
}

func getPersistentVolumes(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	persistentvolumesFolderPath := currentContextPath + "/cluster-scoped-resources/core/persistentvolumes/"
	_persistentvolumes, _ := ioutil.ReadDir(persistentvolumesFolderPath)

	_headers := []string{"name", "capacity", "access mode", "reclaim policy", "status", "claim", "storageclass", "reason", "age", "volumemode"}
	var data [][]string

	_PersistentVolumesList := PersistentVolumesItems{ApiVersion: "v1"}
	for _, f := range _persistentvolumes {
		persistentvolumeYamlPath := persistentvolumesFolderPath + f.Name()
		_file := helpers.ReadYaml(persistentvolumeYamlPath)
		PersistentVolume := corev1.PersistentVolume{}
		if err := yaml.Unmarshal([]byte(_file), &PersistentVolume); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + persistentvolumeYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != PersistentVolume.Name {
			continue
		}

		if outputFlag == "yaml" {
			_PersistentVolumesList.Items = append(_PersistentVolumesList.Items, PersistentVolume)
			continue
		}

		if outputFlag == "json" {
			_PersistentVolumesList.Items = append(_PersistentVolumesList.Items, PersistentVolume)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_PersistentVolumesList.Items = append(_PersistentVolumesList.Items, PersistentVolume)
			continue
		}
		// CAPACITY
		capacity := PersistentVolume.Spec.Capacity.Storage().String()
		// ACCESS MODE
		accessMode := ""
		for _, am := range PersistentVolume.Spec.AccessModes {
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
		//reclaim policy
		reclaimPolicy := string(PersistentVolume.Spec.PersistentVolumeReclaimPolicy)
		//STATUS
		status := string(PersistentVolume.Status.Phase)

		//CLAIM
		claim := ""
		if PersistentVolume.Spec.ClaimRef == nil {
			claim = ""
		} else {
			claim = string(PersistentVolume.Spec.ClaimRef.Namespace) + "/" + string(PersistentVolume.Spec.ClaimRef.Name)
		}

		//SC
		sc := PersistentVolume.Spec.StorageClassName

		//REASON
		reason := PersistentVolume.Status.Reason

		//VOLUME MODE
		volumeMode := "Filesystem"
		if string(*PersistentVolume.Spec.VolumeMode) == "Block" {
			volumeMode = "Block"
		}

		//AGE
		age := helpers.GetAge(persistentvolumeYamlPath, PersistentVolume.GetCreationTimestamp())

		labels := helpers.ExtractLabels(PersistentVolume.GetLabels())
		_list := []string{PersistentVolume.Name, capacity, accessMode, reclaimPolicy, status, claim, sc, reason, age, volumeMode}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 9, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:9] // -A
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
		resource = _PersistentVolumesList.Items[0]
	} else {
		resource = _PersistentVolumesList
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

var PersistentVolume = &cobra.Command{
	Use:     "persistentvolume",
	Aliases: []string{"persistentvolume", "pv"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getPersistentVolumes(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"omc/cmd/helpers"
	"omc/models"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type PodsItems struct {
	ApiVersion string        `json:"apiVersion"`
	Items      []*corev1.Pod `json:"items"`
}

func getPods(omcConfigFile string, aNamespacesFlag bool) {
	headers := []string{"namespace", "name", "ready", "status", "restarts", "age"}
	file, _ := ioutil.ReadFile(omcConfigFile)
	omcConfigJson := models.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)
	// get CurrentContext
	var CurrentContext models.Context
	var DefaultConfigNamespace string
	var contexts []models.Context
	contexts = omcConfigJson.Contexts
	for _, context := range contexts {
		if context.Current == "*" {
			CurrentContext = context
			DefaultConfigNamespace = context.Project
			break
		}
	}
	// get quay-io-... string
	files, err := ioutil.ReadDir(CurrentContext.Path)
	if err != nil {
		log.Fatal(err)
	}
	var QuayString string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "quay") {
			QuayString = f.Name()
			break
		}
	}
	if QuayString == "" {
		log.Fatal("Some error occurred, wrong must-gather file composition")
	}
	var namespaces []string
	if aNamespacesFlag == true {
		_namespaces, _ := ioutil.ReadDir(CurrentContext.Path + "/" + QuayString + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	}
	if namespace != "" && !aNamespacesFlag {
		var _namespace = namespace
		namespaces = append(namespaces, _namespace)
	}
	if namespace == "" && !aNamespacesFlag {
		var _namespace = DefaultConfigNamespace
		namespaces = append(namespaces, _namespace)
	}

	var data [][]string

	for _, _namespace := range namespaces {
		var _Items PodsItems
		CurrentNamespacePath := CurrentContext.Path + "/" + QuayString + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/pods.yaml")
		if err != nil && !aNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/pods.yaml")
			os.Exit(1)
		}
		for _, Pod := range _Items.Items {
			// pod path

			var containers string
			if len(Pod.Spec.Containers) != 0 {
				containers = strconv.Itoa(len(Pod.Spec.Containers))
			} else {
				containers = "0"
			}
			var containerStatuses = Pod.Status.ContainerStatuses // DA VALIDARE L'ESISTENZA
			// ready
			containers_ready := 0
			for _, i := range containerStatuses {
				if i.Ready == true {
					containers_ready = containers_ready + 1
				}
			}
			// restarts
			ContainersRestarts := 0
			for _, i := range containerStatuses {
				if int(i.RestartCount) > ContainersRestarts {
					ContainersRestarts = int(i.RestartCount)
				}
			}
			ContainersReady := strconv.Itoa(containers_ready) + "/" + containers
			//age
			PodsFile, err := os.Stat(CurrentNamespacePath + "/core/pods.yaml")

			if err != nil {
				fmt.Println(err)
			}
			// check podfile last time modification as t2
			t2 := PodsFile.ModTime()
			layout := "2006-01-02 15:04:05 -0700 MST"
			t1, _ := time.Parse(layout, Pod.ObjectMeta.CreationTimestamp.String())
			diffTime := t2.Sub(t1).String()
			d, _ := time.ParseDuration(diffTime)
			diffTimeString := helpers.FormatDiffTime(d)
			//return
			_list := []string{Pod.Namespace, Pod.Name, ContainersReady, string(Pod.Status.Phase), strconv.Itoa(ContainersRestarts), diffTimeString}
			if aNamespacesFlag == true {
				data = append(data, _list)
			} else {
				data = append(data, _list[1:])
			}
		}
	}
	if aNamespacesFlag == true {
		helpers.PrintTable(headers, data)
	} else {
		helpers.PrintTable(headers[1:], data)
	}

}

// podsCmd represents the pods command
var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "pod",
	Run: func(cmd *cobra.Command, args []string) {
		var aNamespacesFlag bool
		aNamespacesFlag, _ = cmd.Flags().GetBool("all-namespaces")
		getPods(viper.ConfigFileUsed(), aNamespacesFlag)
	},
}
var podCmd = &cobra.Command{
	Use:   "pod",
	Short: "alias for pods",
	Run: func(cmd *cobra.Command, args []string) {
		podsCmd.Run(cmd, args)
	},
}

func init() {
	getCmd.AddCommand(podsCmd)
	getCmd.AddCommand(podCmd)
	podsCmd.Flags().BoolP("all-namespaces", "A", false, "all namespaces")
	podCmd.Flags().BoolP("all-namespaces", "A", false, "all namespaces")
}

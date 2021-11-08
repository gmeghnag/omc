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
package logs

import (
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"os"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func logsPods(currentContextPath string, defaultConfigNamespace string, podName string, containerName string, previousFlag bool, allContainersFlag bool) {
	var logFile string
	if previousFlag {
		logFile = "previous.log"
	} else {
		logFile = "current.log"
	}
	var _Items v1.PodList
	CurrentNamespacePath := currentContextPath + "/namespaces/" + defaultConfigNamespace
	_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/pods.yaml")
	if err != nil {
		fmt.Println("error: namespace " + defaultConfigNamespace + " not found.")
		os.Exit(1)
	}
	if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
		fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/core/pods.yaml")
		os.Exit(1)
	}
	podMatch := ""
	for _, Pod := range _Items.Items {
		if podName != Pod.Name {
			continue
		}
		podMatch = podName
		var containers []string
		containerMatch := ""
		if len(Pod.Spec.Containers) == 1 && containerName == "" {
			containerMatch = Pod.Spec.Containers[0].Name
		} else {
			if allContainersFlag {
				for _, c := range Pod.Spec.Containers {
					helpers.Cat(CurrentNamespacePath + "/pods/" + Pod.Name + "/" + c.Name + "/" + c.Name + "/logs/" + logFile)
				}
				return
			} else {
				for _, c := range Pod.Spec.Containers {
					if containerName == c.Name {
						containerMatch = containerName
						break
					}
					containers = append(containers, c.Name)
				}
			}
		}
		if containerMatch == "" {
			for _, c := range Pod.Spec.InitContainers {
				if containerName == c.Name {
					containerMatch = containerName
					break
				}
			}
		}
		if containerMatch == "" {
			if containerName != "" {
				fmt.Println("error: container " + containerName + " is not valid for pod " + Pod.Name)
			} else {
				fmt.Println("error: a container name must be specified for pod "+Pod.Name+", choose one of:", containers)
			}
		} else {
			//fmt.Println("found :", CurrentNamespacePath+"/pods/"+Pod.Name+"/"+containerMatch+"/"+containerMatch+"/logs/"+logFile)
			helpers.Cat(CurrentNamespacePath + "/pods/" + Pod.Name + "/" + containerMatch + "/" + containerMatch + "/logs/" + logFile)
		}
	}
	if podMatch == "" {
		fmt.Println("error: pods " + podName + " not found")
	}
}

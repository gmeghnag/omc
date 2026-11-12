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
	"io"
	"os"

	"github.com/gmeghnag/omc/cmd/helpers"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func findLogDir(namespacePath, podName, containerName string) string {
	traditional := namespacePath + "/pods/" + podName + "/" + containerName + "/" + containerName + "/logs"
	if _, err := os.Stat(traditional); err == nil {
		return traditional
	}
	corePods := namespacePath + "/core/pods/" + podName + "/" + containerName + "/logs"
	if _, err := os.Stat(corePods); err == nil {
		return corePods
	}
	return traditional
}

func logsPods(stdout io.Writer, currentContextPath string, defaultConfigNamespace string, podName string, containerName string, previousFlag bool, rotatedFlag bool, allContainersFlag bool, logLevels []string, insecureFlag bool, tail int64) error {
	var logFilter logLineFilter = NewCRILogFilter(logLevels, nil)
	var _Items v1.PodList
	CurrentNamespacePath := currentContextPath + "/namespaces/" + defaultConfigNamespace
	// helpers.PodObjects resolves every pod storage layout a must-gather may use
	// (aggregated core/pods.yaml, nested pods/<pod>/<pod>.yaml, flat
	// core/pods/<pod>.yaml) and, unlike reading core/pods.yaml directly, treats
	// an empty aggregated placeholder as absent instead of as "no such pod".
	objects, oErr := helpers.PodObjects(CurrentNamespacePath)
	if oErr != nil {
		return oErr
	}
	for _, object := range objects {
		var pod v1.Pod
		if err := yaml.Unmarshal(object.Data, &pod); err != nil {
			return fmt.Errorf("error unmarshaling %s: %w", object.Path, err)
		}
		_Items.Items = append(_Items.Items, pod)
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
					log := NewLogReader(findLogDir(CurrentNamespacePath, Pod.Name, c.Name))
					log.WithFilter(logFilter)
					log.WithTail(tail)
					if previousFlag {
						log.FromPrevious()
					}
					if rotatedFlag {
						log.FromRotated()
					}
					if insecureFlag {
						log.FromInsecure()
					}
					if err := log.Read(stdout); err != nil {
						return err
					}
				}
				return nil
			} else {
				var containerSlice []v1.Container
				containerSlice = append(containerSlice, Pod.Spec.Containers...)
				containerSlice = append(containerSlice, Pod.Spec.InitContainers...)
				for _, c := range containerSlice {
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
				return fmt.Errorf("container %s is not valid for pod %s", containerName, Pod.Name)
			} else {
				return fmt.Errorf("a container name must be specified for pod %s, choose one of: %v", Pod.Name, containers)
			}
		} else {
			log := NewLogReader(findLogDir(CurrentNamespacePath, Pod.Name, containerMatch))
			log.WithFilter(logFilter)
			log.WithTail(tail)
			if previousFlag {
				log.FromPrevious()
			}
			if rotatedFlag {
				log.FromRotated()
			}
			if insecureFlag {
				log.FromInsecure()
			}
			if err := log.Read(stdout); err != nil {
				return err
			}
		}
	}
	if podMatch == "" {
		return fmt.Errorf("pods %s not found", podName)
	}
	return nil
}

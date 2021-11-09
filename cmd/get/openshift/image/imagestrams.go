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
package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/vars"
	"os"
	"strings"

	v1 "github.com/openshift/api/image/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type ImageStreamsItems struct {
	ApiVersion string            `json:"apiVersion"`
	Items      []*v1.ImageStream `json:"items"`
}

func GetImageStreams(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string, allResources bool) bool {
	_headers := []string{"namespace", "name", "image repository", "tags", "updated"}

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
	var _ImageStreamsList = ImageStreamsItems{ApiVersion: "v1"}
	for _, _namespace := range namespaces {
		var _Items ImageStreamsItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/image.openshift.io/imagestreams.yaml")
		if err != nil && !allNamespacesFlag {
			fmt.Println("No resources found in " + _namespace + " namespace.")
			os.Exit(1)
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Println("Error when trying to unmarshall file " + CurrentNamespacePath + "/image.openshift.io/imagestreams.yaml")
			os.Exit(1)
		}

		for _, ImageStream := range _Items.Items {
			if resourceName != "" && resourceName != ImageStream.Name {
				continue
			}

			if outputFlag == "yaml" {
				_ImageStreamsList.Items = append(_ImageStreamsList.Items, ImageStream)
				continue
			}

			if outputFlag == "json" {
				_ImageStreamsList.Items = append(_ImageStreamsList.Items, ImageStream)
				continue
			}

			if strings.HasPrefix(outputFlag, "jsonpath=") {
				_ImageStreamsList.Items = append(_ImageStreamsList.Items, ImageStream)
				continue
			}

			//name
			ImageStreamName := ImageStream.Name
			if allResources {
				ImageStreamName = "imagestream.image.openshift.io/" + ImageStreamName
			}
			//image repository
			imageRepository := ImageStream.Status.DockerImageRepository
			//tags
			tags := ""
			for _, tag := range ImageStream.Status.Tags {
				tags += tag.Tag + ","
			}
			tags = strings.TrimRight(tags, ",")
			//updated
			updated := helpers.GetAge(CurrentNamespacePath+"/image.openshift.io/imagestreams.yaml", ImageStream.GetCreationTimestamp())
			//labels
			labels := helpers.ExtractLabels(ImageStream.GetLabels())
			_list := []string{ImageStream.Namespace, ImageStreamName, imageRepository, tags, updated}
			data = helpers.GetData(data, allNamespacesFlag, showLabels, labels, outputFlag, 5, _list)

			if resourceName != "" && resourceName == ImageStreamName {
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
			headers = _headers[0:5]
		} else {
			headers = _headers[1:5]
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

	if len(_ImageStreamsList.Items) == 0 {
		if !allResources {
			fmt.Println("No resources found in " + namespace + " namespace.")
		}
		return true
	}

	var resource interface{}
	if resourceName != "" {
		resource = _ImageStreamsList.Items[0]
	} else {
		resource = _ImageStreamsList
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

var ImageStream = &cobra.Command{
	Use:     "imagestream",
	Aliases: []string{"imagestreams", "is", "imagestream.imagestream.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		GetImageStreams(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, false)
	},
}

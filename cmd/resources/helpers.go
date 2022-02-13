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
package resources

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"sigs.k8s.io/yaml"
)

func getResources(output string) error {
	resp, err := http.Get("https://github.com/gmeghnag/omc/blob/main/api-resources.yaml")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var resourceList ResourceList
	//Decode the data
	body := resp.Body
	bodyBytes, _ := io.ReadAll(body)
	if err := yaml.Unmarshal(bodyBytes, &resourceList); err != nil {
		return err
	}
	var data [][]string
	headers := []string{"name", "shortnames", "apiversion", "namespaced", "kind"}
	if output == "wide" {
		headers = append(headers, "since")
	}
	for _, resource := range resourceList.Resources {
		shortNames := strings.Join(resource.ShortNames, ",")
		_list := []string{resource.Name, shortNames, resource.ApiVersion, fmt.Sprint(resource.Namespaced), resource.Kind}
		if output == "wide" {
			_list = append(_list, resource.SupportedSince)
		}
		data = append(data, _list)
	}
	helpers.PrintTable(headers, data)
	return nil
}

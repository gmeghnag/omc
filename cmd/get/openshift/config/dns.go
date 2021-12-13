/*
Copyright © 2021 Bram Verschueren <bverschueren@redhat.com>

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
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

type dnsesItems struct {
	ApiVersion string         `json:"apiVersion"`
	Items      []configv1.DNS `json:"items"`
}

func getDNS(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	dnsesYamlPath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/dnses.yaml"

	_headers := []string{"name", "age"}
	var data [][]string

	_file, _ := ioutil.ReadFile(dnsesYamlPath)
	DNSList := configv1.DNSList{}
	if err := yaml.Unmarshal([]byte(_file), &DNSList); err != nil {
		fmt.Println("Error when trying to unmarshal file: " + dnsesYamlPath)
		os.Exit(1)
	}

	_dnsesList := dnsesItems{ApiVersion: "v1"}

	for _, DNS := range DNSList.Items {

		if resourceName != "" && resourceName != DNS.Name {
			continue
		}

		_dnsesList.Items = append(_dnsesList.Items, DNS)

		DNSName := DNS.Name
		since := helpers.GetAge(dnsesYamlPath, DNS.GetCreationTimestamp())
		labels := helpers.ExtractLabels(DNS.GetLabels())
		_list := []string{DNSName, since}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:2] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)

	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
	}
	var resource interface{}
	if resourceName != "" {
		resource = _dnsesList.Items[0]
	} else {
		resource = _dnsesList
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

var DNS = &cobra.Command{
	Use:     "dns",
	Aliases: []string{"dnses", "dns.config.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getDNS(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

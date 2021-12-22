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
package operator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/duration"

	"sigs.k8s.io/yaml"
)

type authenticationItems struct {
	ApiVersion string                      `json:"apiVersion"`
	Items      []operatorv1.Authentication `json:"items"`
}

func getAuthentication(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	// There is only one authentication per cluster, therefore the must gather
	// only contains a single authentication rather than a list. Do not take
	// this file as an example because most of what you see is exceptional.

	if resourceName != "" && resourceName != "cluster" {
		fmt.Println("omc only supports the \"cluster\" authentication. Try omc get authentication or omc get authentication cluster")
		os.Exit(1)
	}

	authenticationYamlPath := currentContextPath + "/cluster-scoped-resources/operator.openshift.io/authentications/cluster.yaml"

	_file, _ := ioutil.ReadFile(authenticationYamlPath)

	authentication := operatorv1.Authentication{}
	if err := yaml.Unmarshal([]byte(_file), &authentication); err != nil {
		fmt.Println("Error when trying to unmarshal file: " + authenticationYamlPath)
		os.Exit(1)
	}

	if outputFlag == "" || outputFlag == "wide" {
		headers := []string{"NAME", "AGE"}
		authenticationName := authentication.Name
		age := ""
		//if authentication.CreationTimestamp != nil {
		age = duration.ShortHumanDuration(time.Now().Sub(authentication.CreationTimestamp))
		//age = authentication.CreationTimestamp.String()
		//}
		authenticationAGE := age
		labels := helpers.ExtractLabels(authentication.GetLabels())

		cn := []string{authenticationName, authenticationAGE}

		if showLabels {
			headers = append(headers, "labels")
			cn = append(cn, labels)
		}

		data := [][]string{cn}
		helpers.PrintTable(headers, data)
	}

	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(authentication)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(authentication, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(authentication, jsonPathTemplate)
	}
	return false
}

var Authentication = &cobra.Command{
	Use:     "authentication.operator",
	Aliases: []string{"authentication.operator", "authentication.operator.openshift.io", "authentications.operator", "authentications.operator.openshift.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getAuthentication(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}

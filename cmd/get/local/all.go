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
package local

import (
	"fmt"

	"github.com/gmeghnag/omc/cmd/get/apps"
	"github.com/gmeghnag/omc/cmd/get/batch"
	"github.com/gmeghnag/omc/cmd/get/core"
	appz "github.com/gmeghnag/omc/cmd/get/openshift/apps"
	"github.com/gmeghnag/omc/cmd/get/openshift/build"
	"github.com/gmeghnag/omc/cmd/get/openshift/image"
	"github.com/gmeghnag/omc/cmd/get/openshift/route"
	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

var All = &cobra.Command{
	Use: "all",
	Run: func(cmd *cobra.Command, args []string) {
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		empty := core.GetPods(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = core.GetReplicationControllers(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = core.GetServices(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = apps.GetDaemonSets(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = apps.GetDeployments(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = apps.GetReplicaSets(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = batch.GetJobs(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = appz.GetDeploymentConfigs(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = build.GetBuildConfigs(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = build.GetBuilds(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = image.GetImageStreams(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
		empty = route.GetRoutes(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate, true)
		if !empty {
			fmt.Println("")
		}
	},
	Hidden: true,
}

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
package get

import (
	"fmt"

	"github.com/gmeghnag/omc/cmd/get/apps"
	"github.com/gmeghnag/omc/cmd/get/batch"
	"github.com/gmeghnag/omc/cmd/get/certificate"
	"github.com/gmeghnag/omc/cmd/get/core"
	"github.com/gmeghnag/omc/cmd/get/istio/networking"
	"github.com/gmeghnag/omc/cmd/get/local"
	"github.com/gmeghnag/omc/cmd/get/maistra"
	appz "github.com/gmeghnag/omc/cmd/get/openshift/apps"
	"github.com/gmeghnag/omc/cmd/get/openshift/build"
	"github.com/gmeghnag/omc/cmd/get/openshift/config"
	"github.com/gmeghnag/omc/cmd/get/openshift/image"
	"github.com/gmeghnag/omc/cmd/get/openshift/machine"
	"github.com/gmeghnag/omc/cmd/get/openshift/machineconfiguration"
	"github.com/gmeghnag/omc/cmd/get/openshift/network"
	"github.com/gmeghnag/omc/cmd/get/openshift/operator"
	"github.com/gmeghnag/omc/cmd/get/openshift/route"

	"os"
	"strings"

	operators "github.com/gmeghnag/omc/cmd/get/operator-framework"
	"github.com/gmeghnag/omc/cmd/get/storage"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

var outputStringVar string
var allNamespaceBoolVar, showLabelsBoolVar bool

var GetCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		// support (not completed) for comma separated resources
		if len(args) == 1 && strings.Contains(args[0], ",") {
			commaSeparatedResources := strings.TrimSuffix(args[0], ",")
			commaSeparatedResources = strings.TrimPrefix(commaSeparatedResources, ",")
			resources := strings.Split(commaSeparatedResources, ",")
			for _, r := range resources {
				c, _, err := cmd.Find([]string{r})
				if err != nil {
					fmt.Println("err", err.Error())
				}
				os.Args = append([]string{os.Args[0], "get", r}, os.Args[3:]...)
				c.Execute()
				fmt.Println("")
			}
			os.Exit(0)
		}
		// object not found
		fmt.Println("Invalid object type:", args[0])
		os.Exit(1)
	},
}

func init() {
	if len(os.Args) > 2 && os.Args[1] == "get" {
		os.Args[2] = strings.ToLower(os.Args[2])
		if strings.Contains(os.Args[2], "/") {
			seg := strings.Split(os.Args[2], "/")
			resource, name := seg[0], seg[1]
			os.Args = append([]string{os.Args[0], "get", resource, name}, os.Args[3:]...)
		}
	}
	GetCmd.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	GetCmd.PersistentFlags().BoolVarP(&vars.ShowLabelsBoolVar, "show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	GetCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")
	GetCmd.PersistentFlags().StringVarP(&vars.LabelSelectorStringVar, "selector", "l", "", "selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	GetCmd.AddCommand(
		apps.DaemonSet,
		apps.Deployment,
		apps.ReplicaSet,
		appz.DeploymentConfig,
		batch.CronJob,
		batch.Job,
		build.Build,
		build.BuildConfig,
		certificate.CertificateSigningRequest,
		config.ClusterOperator,
		config.ClusterVersion,
		config.Proxy,
		config.Infrastructure,
		config.Network,
		config.DNS,
		core.ConfigMap,
		core.Endpoint,
		core.Event,
		core.Namespace,
		core.Node,
		core.PersistentVolume,
		core.PersistentVolumeClaim,
		core.Pod,
		core.ReplicationController,
		core.Secret,
		core.Service,
		image.ImageStream,
		machine.Machine,
		machine.MachineSet,
		machineconfiguration.MachineConfig,
		machineconfiguration.MachineConfigPool,
		maistra.ServiceMeshControlPlane,
		maistra.ServiceMeshMemberRoll,
		networking.DestinationRule,
		networking.Gateway,
		networking.VirtualService,
		local.All,
		local.MustGather,
		operators.ClusterServiceVersion,
		operators.InstallPlan,
		operators.Subscription,
		route.Route,
		storage.StorageClass,
		network.ClusterNetwork,
		operator.Authentication,
	)
}

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

	"github.com/gmeghnag/omc/cmd/get/apiextensions"
	"github.com/gmeghnag/omc/cmd/get/apps"
	"github.com/gmeghnag/omc/cmd/get/batch"
	"github.com/gmeghnag/omc/cmd/get/certificate"
	"github.com/gmeghnag/omc/cmd/get/core"
	"github.com/gmeghnag/omc/cmd/get/istio/networking"
	"github.com/gmeghnag/omc/cmd/get/local"
	"github.com/gmeghnag/omc/cmd/get/maistra"
	baremetalhost "github.com/gmeghnag/omc/cmd/get/metal3"
	networkingv1 "github.com/gmeghnag/omc/cmd/get/networking"
	"github.com/gmeghnag/omc/cmd/get/openshift/apiserver"
	appz "github.com/gmeghnag/omc/cmd/get/openshift/apps"
	"github.com/gmeghnag/omc/cmd/get/openshift/build"
	"github.com/gmeghnag/omc/cmd/get/openshift/config"
	"github.com/gmeghnag/omc/cmd/get/openshift/image"
	"github.com/gmeghnag/omc/cmd/get/openshift/logging"
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
var emptyslice []string
var resourcesAndObjects [][]string

var GetCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		// support (not completed) for comma separated resources
		if strings.Contains(args[0], "/") {
			for _, resourceAndObjectArg := range args {
				resourceAndObjectArg = strings.ToLower(resourceAndObjectArg)
				if len(strings.Split(resourceAndObjectArg, "/")) != 2 {
					fmt.Println("error: there is no need to specify a resource type as a separate argument when passing arguments in resource/name form (e.g. 'oc get resource/<resource_name>' instead of 'oc get resource resource/<resource_name>'")
					os.Exit(1)
				}
				seg := strings.Split(resourceAndObjectArg, "/")
				crdAlias, objectName := seg[0], seg[1]
				aliasAlreadyPresentInResourcesAndObjects := false
				for i, resourceAndObject := range resourcesAndObjects {
					if crdAlias == resourceAndObject[0] {
						aliasAlreadyPresentInResourcesAndObjects = true
						resourcesAndObjects[i] = append(resourcesAndObjects[i], objectName)
						break
					}
				}
				if !aliasAlreadyPresentInResourcesAndObjects {
					resourcesAndObjects = append(resourcesAndObjects, []string{crdAlias, objectName})
				}

			}
			for _, resourceAndObject := range resourcesAndObjects {
				c, _, err := cmd.Find([]string{resourceAndObject[0]})
				if err != nil {
					fmt.Println("err", err.Error())
					os.Exit(1)
				}
				if c.Use != "get" {
					os.Args = append([]string{os.Args[0], "get", resourceAndObject[0]}, resourceAndObject[1:]...)
					c.Execute()
				} else {
					isValidResource := getGenericResourceFromCRD(resourceAndObject[0], resourceAndObject[1:])
					if !isValidResource {
						fmt.Println("Invalid object type:", resourceAndObject[0])
						os.Exit(1)
					}
				}
			}
		}
		if len(args) == 1 && !strings.Contains(args[0], "/") {
			if strings.Contains(args[0], ",") {
				commaSeparatedResources := strings.TrimSuffix(args[0], ",")
				commaSeparatedResources = strings.TrimPrefix(commaSeparatedResources, ",")
				crdAliases := strings.Split(commaSeparatedResources, ",")
				for _, crdAlias := range crdAliases {
					c, _, err := cmd.Find([]string{crdAlias})
					if err != nil {
						fmt.Println("err", err.Error())
						os.Exit(1)
					}
					if c.Use != "get" {
						os.Args = append([]string{os.Args[0], "get", crdAlias}, "")
						c.Execute()
						fmt.Println("")
					} else {
						isValidResource := getGenericResourceFromCRD(crdAlias, emptyslice)
						if !isValidResource {
							fmt.Println("Invalid object type:", crdAlias)
							os.Exit(1)
						}
						fmt.Println("")
					}
				}
			} else {
				c, _, err := cmd.Find([]string{args[0]})
				if err != nil {
					fmt.Println("err", err.Error())
					os.Exit(1)
				}
				if c.Use != "get" {
					os.Args = append([]string{os.Args[0], "get", args[0]}, "")
					c.Execute()
				} else {
					isValidResource := getGenericResourceFromCRD(args[0], emptyslice)
					if !isValidResource {
						fmt.Println("Invalid object type:", args[0])
						os.Exit(1)
					}
				}
			}
		} else if len(args) > 1 {
			c, _, err := cmd.Find([]string{args[0]})
			if err != nil {
				fmt.Println("err", err.Error())
				os.Exit(1)
			}
			if c.Use != "get" {
				os.Args = append([]string{os.Args[0], "get", args[0]}, args[1:]...)
				c.Execute()
			} else {
				isValidResource := getGenericResourceFromCRD(args[0], args[1:])
				if !isValidResource {
					fmt.Println("Invalid object type:", args[0])
					os.Exit(1)
				}
			}
		}
	},
}

func init() {
	GetCmd.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	GetCmd.PersistentFlags().BoolVarP(&vars.ShowLabelsBoolVar, "show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	GetCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")
	GetCmd.PersistentFlags().StringVarP(&vars.LabelSelectorStringVar, "selector", "l", "", "selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	GetCmd.AddCommand(
		apiextensions.CustomResourceDefinition,
		apiserver.APIRequestCount,
		apps.DaemonSet,
		apps.Deployment,
		apps.ReplicaSet,
		apps.StatefulSet,
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
		baremetalhost.BareMetalHost,
		maistra.ServiceMeshControlPlane,
		maistra.ServiceMeshMemberRoll,
		networkingv1.Ingress,
		networkingv1.IngressController,
		networking.DestinationRule,
		networking.Gateway,
		networking.VirtualService,
		local.All,
		local.MustGather,
		logging.ClusterLogForwarder,
		logging.ClusterLogging,
		operators.ClusterServiceVersion,
		operators.InstallPlan,
		operators.Subscription,
		route.Route,
		storage.StorageClass,
		network.ClusterNetwork,
		operator.Authentication,
	)
}

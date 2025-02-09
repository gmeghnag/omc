/*
Copyright © 2024 Bram Verschueren <bverschueren@redhat.com>

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
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/bverschueren/in2un/pkg/deserializer"
	"github.com/bverschueren/in2un/pkg/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

var OverrideApiVersion, OverrideKind, Output string

var getCmd = &cobra.Command{
	Use:  "get",
	Args: cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if AllNamespaces {
			Namespace = "_all_" // TODO: export and use global (?) AllNamespaceValue
		}
	},

	Short: "Parse Insights data as generic unstructured (https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured) data.",
	Run: func(cmd *cobra.Command, args []string) {
		resourceGroup, resourceName := processArgs(args)
		ir, err := reader.NewInsightsReader(viper.GetString("active"))
		if err != nil {
			log.Fatal(err)
		}
		found := ir.ReadResource(resourceGroup, resourceName, Namespace, OverrideApiVersion, OverrideKind)
		handleOutput(Output, found)
	},
}

func handleOutput(format string, obj *unstructured.UnstructuredList) {
	if hasDummyFields(obj) {
		log.Warning("Hint: use --api-version and --kind to override dummy values for missing fields in insights archives")
	}
	var printr printers.ResourcePrinter
	switch format {
	case "yaml":
		printr = printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.YAMLPrinter{})
		if err := printr.PrintObj(obj, os.Stdout); err != nil {
			panic(err.Error())
		}
	case "json":
		printr = printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.JSONPrinter{})
		if err := printr.PrintObj(obj, os.Stdout); err != nil {
			panic(err.Error())
		}
	case "name":
		printr = printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.NamePrinter{})
		if err := printr.PrintObj(obj, os.Stdout); err != nil {
			panic(err.Error())
		}
	default: //table printer
		printr = printers.NewTypeSetter(scheme.Scheme).ToPrinter(printers.NewTablePrinter(printers.PrintOptions{}))
		if err := printr.PrintObj(obj, os.Stdout); err != nil {
			panic(err.Error())
		}
	}
}

func hasDummyFields(obj *unstructured.UnstructuredList) bool { //TODO: generic warning loop interface
	if len(obj.Items) > 0 {
		return obj.Items[0].Object["apiVersion"] == deserializer.MissingTypeMetaFieldValue || obj.Items[0].Object["kind"] == deserializer.MissingTypeMetaFieldValue
	} else {
		return false
	}
}

func init() {
	InsightsCmd.AddCommand(getCmd)
	//getCmd.PersistentFlags().BoolVarP(&AllNamespaces, "all-namespaces", "A", false, "Set the namespace scope for this CLI request to all namespaces")
	getCmd.Flags().BoolVarP(&AllNamespaces, "all-namespaces", "A", false, "Set the namespace scope for this CLI request to all namespaces")
	getCmd.Flags().StringVarP(&Output, "output", "o", "table", "Output format. One of: (json, yaml, name).")
	getCmd.Flags().StringVar(&OverrideApiVersion, "api-version", "", "Override the apiVersion for the specified resource. By default the apiVersion is trimmed off resource in insights data")
	getCmd.Flags().StringVar(&OverrideKind, "kind", "", "Override the apiVersion for the specified resource. By default the apiVersion is trimmed off resource in insights data")
}

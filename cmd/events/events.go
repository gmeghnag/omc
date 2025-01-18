/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

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
package events

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gmeghnag/omc/cmd/get"
	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliprint "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/klog/v2"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/printers"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
)

var EventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Display events that are sorted by time.",
	Run: func(cmd *cobra.Command, args []string) {
		err := Validate()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		eventList := GetEventList(vars.MustGatherRootPath, vars.Namespace, vars.AllNamespaceBoolVar)
		FilterEventList(&eventList, vars.EventTypes, vars.ForResource)
		SortEventList(&eventList)
		PrintEventList(&eventList, vars.MustGatherRootPath, vars.OutputStringVar, vars.Namespace, vars.AllNamespaceBoolVar)
	},
}

func init() {
	EventsCmd.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	EventsCmd.PersistentFlags().StringVar(&vars.ForResource, "for", "", "Filter events to only those pertaining to the specified resource.")
	EventsCmd.PersistentFlags().StringSliceVar(&vars.EventTypes, "types", vars.EventTypes, "Output only events of given types.")
	EventsCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|name")
}

func Validate() error {
	if len(vars.ForResource) > 0 && !strings.Contains(vars.ForResource, "/") {
		return fmt.Errorf("error when parsing --for: resource must be in resource/name form")
	}
	return nil
}

func GetEventList(context string, selectedNs string, allNamespaces bool) (eventList corev1.EventList) {
	eventsLocation := "/core/events.yaml"
	nsFolder := context + "/namespaces/"
	var namespaces []string
	if allNamespaces {
		fileObj, err := os.Open(nsFolder)
		if err != nil {
			klog.ErrorS(err, "Unable to read "+nsFolder)
			os.Exit(1)
		}
		namespaces, err = fileObj.Readdirnames(0)
		if err != nil {
			klog.ErrorS(err, "Unable to list directories in "+nsFolder)
			os.Exit(1)
		}
	} else {
		namespaces = append(namespaces, selectedNs)
	}

	for _, namespace := range namespaces {
		eventsPath := nsFolder + namespace + eventsLocation
		eventsFile, err := os.ReadFile(eventsPath)
		if err != nil {
			klog.V(5).ErrorS(err, "Unable to read events.yaml")
			continue
		}
		var nsEvents corev1.EventList
		if err := yaml.Unmarshal([]byte(eventsFile), &nsEvents); err != nil {
			klog.V(3).ErrorS(err, "Unable to parse Kubernetes EventList object from "+eventsPath)
			continue
		}

		if eventList.Kind == "" {
			eventList.SetGroupVersionKind(nsEvents.GroupVersionKind())
		}
		eventList.Items = append(eventList.Items, nsEvents.Items...)
	}
	return eventList
}

func FilterEventList(eventList *corev1.EventList, types []string, forResource string) {
	if len(types) > 0 {
		var filteredType []corev1.Event
		for _, event := range eventList.Items {
			for _, _type := range types {
				if strings.EqualFold(event.Type, _type) {
					filteredType = append(filteredType, event)
				}
			}
		}
		eventList.Items = filteredType
	}

	if len(forResource) > 0 {
		splitStr := strings.Split(forResource, "/")
		inputResource, resourceName := splitStr[0], splitStr[1]
		inputResource = strings.ToLower(inputResource)
		_, resourceGroup, resourceKind, _, err := get.KindGroupNamespaced(inputResource)
		if err != nil {
			klog.V(3).ErrorS(err, "Couldn't get --for resource")
		}

		var filteredFor []corev1.Event
		for _, event := range eventList.Items {
			involved := event.InvolvedObject
			gvk := involved.GroupVersionKind()
			if gvk.Group == "" {
				gvk.Group = "core"
			}
			if strings.EqualFold(gvk.Group, resourceGroup) && strings.EqualFold(gvk.Kind, resourceKind) && strings.EqualFold(involved.Name, resourceName) {
				filteredFor = append(filteredFor, event)
			}
		}
		eventList.Items = filteredFor
	}
}

func SortEventList(eventList *corev1.EventList) {
	events := eventList.Items
	slices.SortFunc(events, func(i, j corev1.Event) int {
		return GetLastTime(i).Time.Compare(GetLastTime(j).Time)
	})
}

func PrintEventList(eventList *corev1.EventList, context string, output string, selectedNs string, allNamespaces bool) {
	if len(eventList.Items) == 0 {
		if allNamespaces {
			fmt.Printf("No events found.\n")
		} else {
			fmt.Printf("No events found in %s namespace.\n", selectedNs)
		}
	}

	var cliPrinter cliprint.ResourcePrinter
	if output == "name" {
		cliPrinter := cliprint.NamePrinter{ShortOutput: true}
		for _, event := range eventList.Items {
			err := cliPrinter.PrintObj(&event, os.Stdout)
			if err != nil {
				klog.V(3).ErrorS(err, "Error when outputting names of events")
			}
		}
	} else if output == "yaml" {
		cliPrinter := cliprint.YAMLPrinter{}
		err := cliPrinter.PrintObj(eventList, os.Stdout)
		if err != nil {
			klog.V(3).ErrorS(err, "Error when outputting YAML of events")
		}
	} else if output == "json" {
		cliPrinter := cliprint.JSONPrinter{}
		err := cliPrinter.PrintObj(eventList, os.Stdout)
		if err != nil {
			klog.V(3).ErrorS(err, "Error when outputting JSON of events")
		}
	} else {
		// There is no handler for corev1.EventList in the table generator
		var printList api.EventList
		convertType(eventList, &printList)
		table, err := vars.TableGenerator.GenerateTable(&printList, printers.GenerateOptions{})
		if err != nil {
			klog.V(3).ErrorS(err, "Error when generating table output of events")
		}

		if table.ColumnDefinitions[0].Name == "Last Seen" {
			for i, event := range eventList.Items {
				last := GetLastTime(event)
				age := helpers.GetAge(context, last)
				table.Rows[i].Cells[0] = age
			}
		}

		if allNamespaces {
			table.ColumnDefinitions = append([]metav1.TableColumnDefinition{{Format: "string", Name: "Namespace"}}, table.ColumnDefinitions...)
			for i, event := range eventList.Items {
				table.Rows[i].Cells = append([]interface{}{event.GetNamespace()}, table.Rows[i].Cells...)
			}
		}

		cliPrinter = cliprint.NewTablePrinter(cliprint.PrintOptions{})
		err = cliPrinter.PrintObj(table, os.Stdout)
		if err != nil {
			klog.V(3).ErrorS(err, "Error when outputting table of events")
		}
	}
}

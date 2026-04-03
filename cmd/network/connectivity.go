/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

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
package network

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	operatorcontrolplanev1alpha1 "github.com/openshift/api/operatorcontrolplane/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	podNetworkConnectivityYAML = "/pod_network_connectivity_check/podnetworkconnectivitychecks.yaml"
	msgTruncateLen               = 80
)

var (
	connectivityWide          bool
	connectivityUnhealthyOnly bool
)

// connectivityListEnvelope matches the Kubernetes v1 List shape written by must-gather.
type connectivityListEnvelope struct {
	APIVersion string                                                       `json:"apiVersion"`
	Items      []operatorcontrolplanev1alpha1.PodNetworkConnectivityCheck `json:"items"`
}

var ConnectivityCmd = &cobra.Command{
	Use:   "connectivity",
	Short: "Summarize PodNetworkConnectivityCheck resources from must-gather",
	Long: `Read pod_network_connectivity_check/podnetworkconnectivitychecks.yaml from the active must-gather
and print a concise table of each check (target, reachability, condition details, and last failure from
status.failures when present).

These checks normally live in openshift-network-diagnostics; use --namespace / -n on the root command to
filter if needed. Use --wide for extra columns and untruncated last failure text. Use --unhealthy-only to show checks
where the Reachable condition is not True.

LAST_FAILURE_REASON and LAST_FAILURE_MESSAGE come from the most recent status.failures log entry (empty if
none).`,
	Run: func(cmd *cobra.Command, args []string) {
		if vars.MustGatherRootPath == "" {
			fmt.Fprintln(os.Stderr, "No must-gather context. Run 'omc use <path-to-must-gather>' first.")
			os.Exit(1)
		}
		path := vars.MustGatherRootPath + podNetworkConnectivityYAML
		exist, _ := helpers.Exists(path)
		if !exist {
			fmt.Fprintf(os.Stderr, "Path '"+path+"' does not exist.\n")
			os.Exit(1)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		list, err := loadConnectivityList(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when trying to unmarshal file: %s\n", path)
			os.Exit(1)
		}
		items := list.Items
		slices.SortFunc(items, func(a, b operatorcontrolplanev1alpha1.PodNetworkConnectivityCheck) int {
			return strings.Compare(a.Name, b.Name)
		})
		var headers []string
		if connectivityWide {
			headers = []string{"NAME", "TARGET", "REACHABLE", "LAST_TRANSITION", "LAST_FAILURE_REASON", "LAST_FAILURE_MESSAGE"}
		} else {
			headers = []string{"NAME", "TARGET", "REACHABLE", "LAST_FAILURE_REASON", "LAST_FAILURE_MESSAGE"}
		}
		var data [][]string
		var wantedNamespace string
		if cmd.Root().PersistentFlags().Changed("namespace") {
			wantedNamespace = vars.Namespace
		}
		for _, item := range items {
			if wantedNamespace != "" && item.Namespace != wantedNamespace {
				continue
			}
			if connectivityUnhealthyOnly && isReachableHealthy(item) {
				continue
			}
			_, _, lastTransition := reachableDetails(item)
			lastFailReason, lastFailMsg := latestFailureDetails(item)
			if !connectivityWide {
				if len(lastFailReason) > msgTruncateLen {
					lastFailReason = lastFailReason[:msgTruncateLen-3] + "..."
				}
				if len(lastFailMsg) > msgTruncateLen {
					lastFailMsg = lastFailMsg[:msgTruncateLen-3] + "..."
				}
			}
			var row []string
			if connectivityWide {
				row = []string{
					item.Name,
					item.Spec.TargetEndpoint,
					reachableString(item),
					lastTransition,
					lastFailReason,
					lastFailMsg,
				}
			} else {
				row = []string{
					item.Name,
					item.Spec.TargetEndpoint,
					reachableString(item),
					lastFailReason,
					lastFailMsg,
				}
			}
			data = append(data, row)
		}
		if len(data) == 0 {
			if connectivityUnhealthyOnly {
				fmt.Fprintln(os.Stderr, "No unhealthy PodNetworkConnectivityChecks match the current filters.")
			} else if wantedNamespace != "" {
				fmt.Fprintf(os.Stderr, "No PodNetworkConnectivityChecks found in namespace %q.\n", wantedNamespace)
			} else {
				fmt.Fprintln(os.Stderr, "No PodNetworkConnectivityChecks found in must-gather.")
			}
			os.Exit(0)
		}
		helpers.PrintTable(headers, data)
	},
}

func init() {
	ConnectivityCmd.Flags().BoolVarP(&connectivityWide, "wide", "w", false, "Include full last failure reason and message, and Reachable last transition time")
	ConnectivityCmd.Flags().BoolVar(&connectivityUnhealthyOnly, "unhealthy-only", false, "Only show checks where Reachable is not True")
}

func loadConnectivityList(raw []byte) (*connectivityListEnvelope, error) {
	j, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return nil, err
	}
	var list connectivityListEnvelope
	if err := json.Unmarshal(j, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func reachableString(item operatorcontrolplanev1alpha1.PodNetworkConnectivityCheck) string {
	for _, c := range item.Status.Conditions {
		if c.Type == operatorcontrolplanev1alpha1.Reachable {
			return string(c.Status)
		}
	}
	return "Unknown"
}

func isReachableHealthy(item operatorcontrolplanev1alpha1.PodNetworkConnectivityCheck) bool {
	for _, c := range item.Status.Conditions {
		if c.Type == operatorcontrolplanev1alpha1.Reachable {
			return c.Status == metav1.ConditionTrue
		}
	}
	return false
}

func reachableDetails(item operatorcontrolplanev1alpha1.PodNetworkConnectivityCheck) (reason, message, lastTransition string) {
	for _, c := range item.Status.Conditions {
		if c.Type == operatorcontrolplanev1alpha1.Reachable {
			reason = c.Reason
			message = c.Message
			if !c.LastTransitionTime.IsZero() {
				lastTransition = c.LastTransitionTime.Time.Format("2006-01-02 15:04:05 MST")
			}
			return
		}
	}
	return "", "", ""
}

// latestFailureDetails returns reason and message from the most recent status.failures log entry, or empty strings if none.
func latestFailureDetails(item operatorcontrolplanev1alpha1.PodNetworkConnectivityCheck) (reason, failureMsg string) {
	failures := item.Status.Failures
	if len(failures) == 0 {
		return "", ""
	}
	latest := failures[0]
	for i := 1; i < len(failures); i++ {
		f := failures[i]
		if f.Start.After(latest.Start.Time) {
			latest = f
		}
	}
	return latest.Reason, strings.ReplaceAll(strings.TrimSpace(latest.Message), "\n", " ")
}

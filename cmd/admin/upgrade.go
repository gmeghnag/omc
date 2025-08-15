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
package admin

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/blang/semver"
	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var IncludeNotRecommended bool


func admUpgradeCommand(currentContextPath string) {
	cv := configv1.ClusterVersion{}
	clusterVersionFilePath := currentContextPath + "/cluster-scoped-resources/config.openshift.io/clusterversions/version.yaml"
	clusterversionFilePathExists, _ := helpers.Exists(clusterVersionFilePath)
	if clusterversionFilePathExists {
		_file, _ := os.ReadFile(clusterVersionFilePath)
		if err := yaml.Unmarshal([]byte(_file), &cv); err != nil {
			fmt.Println("Error trying to unmarshal file: " + clusterVersionFilePath)
			os.Exit(1)
		} 
	}
	if cv.Spec.Channel != "" {
		if cv.Spec.Upstream == "" {
			fmt.Fprint(os.Stdout, "Upstream is unset, so the cluster will use an appropriate default.\n")
		} else {
			fmt.Fprintf(os.Stdout, "Upstream: %s\n", cv.Spec.Upstream)
		}
		if len(cv.Status.Desired.Channels) > 0 {
			fmt.Fprintf(os.Stdout, "Channel: %s (available channels: %s)\n", cv.Spec.Channel, strings.Join(cv.Status.Desired.Channels, ", "))
		} else {
			fmt.Fprintf(os.Stdout, "Channel: %s\n", cv.Spec.Channel)
		}
	}
	if len(cv.Status.AvailableUpdates) > 0 {
		fmt.Fprintf(os.Stdout, "\nRecommended updates:\n\n")
			// set the minimal cell width to 14 to have a larger space between the columns for shorter versions
		w := tabwriter.NewWriter(os.Stdout, 14, 2, 1, ' ', 0)
		fmt.Fprintf(w, "  VERSION\tIMAGE\n")
			// TODO: add metadata about version
		sortReleasesBySemanticVersions(cv.Status.AvailableUpdates)
		for _, update := range cv.Status.AvailableUpdates {
			fmt.Fprintf(w, "  %s\t%s\n", update.Version, update.Image)
		}
		w.Flush()
		if c := findClusterOperatorStatusCondition(cv.Status.Conditions, configv1.RetrievedUpdates); c != nil && c.Status == configv1.ConditionFalse {
			fmt.Fprintf(os.Stderr, "warning: Cannot refresh available updates:\n  Reason: %s\n  Message: %s\n\n", c.Reason, strings.ReplaceAll(c.Message, "\n", "\n  "))
		}
	} else {
		if c := findClusterOperatorStatusCondition(cv.Status.Conditions, configv1.RetrievedUpdates); c != nil && c.Status == configv1.ConditionFalse {
			fmt.Fprintf(os.Stderr, "warning: Cannot display available updates:\n  Reason: %s\n  Message: %s\n\n", c.Reason, strings.ReplaceAll(c.Message, "\n", "\n  "))
		} else {
			fmt.Fprintf(os.Stdout, "No updates available. You may still upgrade to a specific release image with --to-image or wait for new updates to be available.\n")
		}
	}

	if IncludeNotRecommended {
		if containsNotRecommendedUpdate(cv.Status.ConditionalUpdates) {
			sortConditionalUpdatesBySemanticVersions(cv.Status.ConditionalUpdates)
			fmt.Fprintf(os.Stdout, "\nUpdates with known issues:\n")
			for _, update := range cv.Status.ConditionalUpdates {
				if c := findCondition(update.Conditions, "Recommended"); c != nil && c.Status != metav1.ConditionTrue {
					fmt.Fprintf(os.Stdout, "\n  Version: %s\n  Image: %s\n", update.Release.Version, update.Release.Image)
					fmt.Fprintf(os.Stdout, "  Reason: %s\n  Message: %s\n", c.Reason, strings.ReplaceAll(strings.TrimSpace(c.Message), "\n", "\n  "))
				}
			}
		} else {
			fmt.Fprintf(os.Stdout, "\nNo updates which are not recommended based on your cluster configuration are available.\n")
		}
    } else if containsNotRecommendedUpdate(cv.Status.ConditionalUpdates) {
		qualifier := ""
		for _, upgrade := range cv.Status.ConditionalUpdates {
			if c := findCondition(upgrade.Conditions, "Recommended"); c != nil && c.Status != metav1.ConditionTrue && c.Status != metav1.ConditionFalse {
				qualifier = fmt.Sprintf(", or where the recommended status is %q,", c.Status)
				break
			}
		}
		fmt.Fprintf(os.Stdout, "\nAdditional updates which are not recommended%s for your cluster configuration are available, to view those re-run the command with --include-not-recommended.\n", qualifier)
	}
}


var Upgrade = &cobra.Command{
	Use:   "upgrade",
	Run: func(cmd *cobra.Command, args []string) {
		admUpgradeCommand(vars.MustGatherRootPath)
	},
}
func init() {
	Upgrade.AddCommand(
		UpgradeRecommend,
	)
	Upgrade.PersistentFlags().BoolVar(&IncludeNotRecommended, "include-not-recommended", false, "Display additional updates which are not recommended based on your cluster configuration.")
}


// sortConditionalUpdatesBySemanticVersions sorts the input slice in decreasing order.
func sortConditionalUpdatesBySemanticVersions(updates []configv1.ConditionalUpdate) {
	sort.Slice(updates, func(i, j int) bool {
		a, errA := semver.Parse(updates[i].Release.Version)
		b, errB := semver.Parse(updates[j].Release.Version)
		if errA == nil && errB != nil {
			return true
		}
		if errB == nil && errA != nil {
			return false
		}
		if errA != nil && errB != nil {
			return updates[i].Release.Version > updates[j].Release.Version
		}
		return a.GT(b)
	})
}


func findCondition(conditions []metav1.Condition, name string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == name {
			return &conditions[i]
		}
	}
	return nil
}
func containsNotRecommendedUpdate(updates []configv1.ConditionalUpdate) bool {
	for _, update := range updates {
		if c := findCondition(update.Conditions, "Recommended"); c != nil && c.Status != metav1.ConditionTrue {
			return true
		}
	}
	return false
}



// sortReleasesBySemanticVersions sorts the input slice in decreasing order.
func sortReleasesBySemanticVersions(versions []configv1.Release) {
	sort.Slice(versions, func(i, j int) bool {
		a, errA := semver.Parse(versions[i].Version)
		b, errB := semver.Parse(versions[j].Version)
		if errA == nil && errB != nil {
			return true
		}
		if errB == nil && errA != nil {
			return false
		}
		if errA != nil && errB != nil {
			return versions[i].Version > versions[j].Version
		}
		return a.GT(b)
	})
}

func findClusterOperatorStatusCondition(conditions []configv1.ClusterOperatorStatusCondition, name configv1.ClusterStatusConditionType) *configv1.ClusterOperatorStatusCondition {
	for i := range conditions {
		if conditions[i].Type == name {
			return &conditions[i]
		}
	}
	return nil
}
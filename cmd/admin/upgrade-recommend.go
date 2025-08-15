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

var showOutdatedReleases, quiet bool

func admUpgradeRecommendCommand(currentContextPath string) error {
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
	majorMinorBuckets := map[uint64]map[uint64][]configv1.ConditionalUpdate{}

	for i, update := range cv.Status.ConditionalUpdates {
		version, err := semver.Parse(update.Release.Version)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "warning: Cannot parse SemVer available update %q: %v", update.Release.Version, err)
			}
			continue
		}

		if minorBuckets := majorMinorBuckets[version.Major]; minorBuckets == nil {
			majorMinorBuckets[version.Major] = make(map[uint64][]configv1.ConditionalUpdate, 0)
		}

		majorMinorBuckets[version.Major][version.Minor] = append(majorMinorBuckets[version.Major][version.Minor], cv.Status.ConditionalUpdates[i])
	}

	for i, update := range cv.Status.AvailableUpdates {
		found := false
		for _, conditionalUpdate := range cv.Status.ConditionalUpdates {
			if conditionalUpdate.Release.Image == update.Image {
				found = true
				break
			}
		}
		if found {
			continue
		}

		version, err := semver.Parse(update.Version)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "warning: Cannot parse SemVer available update %q: %v", update.Version, err)
			}
			continue
		}

		if minorBuckets := majorMinorBuckets[version.Major]; minorBuckets == nil {
			majorMinorBuckets[version.Major] = make(map[uint64][]configv1.ConditionalUpdate, 0)
		}

		majorMinorBuckets[version.Major][version.Minor] = append(majorMinorBuckets[version.Major][version.Minor], configv1.ConditionalUpdate{
			Release: cv.Status.AvailableUpdates[i],
		})
	}

	if c := findClusterOperatorStatusCondition(cv.Status.Conditions, configv1.OperatorUpgradeable); c != nil && c.Status == configv1.ConditionFalse {
		if err := injectUpgradeableAsCondition(cv.Status.Desired.Version, c, majorMinorBuckets); err != nil && !quiet {
			if !quiet {
				fmt.Fprintf(os.Stderr, "warning: Cannot inject %s=%s as a conditional update risk: %s\n\nReason: %s\n  Message: %s\n\n", c.Type, c.Status, err, c.Reason, strings.ReplaceAll(c.Message, "\n", "\n  "))
			}
		}
	}

	//
	//
	

	if len(majorMinorBuckets) == 0 {
		fmt.Fprintf(os.Stdout, "No updates available. You may still upgrade to a specific release image with --to-image or wait for new updates to be available.\n")
		return nil
	}

	majors := make([]uint64, 0, len(majorMinorBuckets))
	for major := range majorMinorBuckets {
		majors = append(majors, major)
	}
	sort.Slice(majors, func(i, j int) bool {
		return majors[i] > majors[j] // sort descending, major updates bring lots of features (enough to justify breaking backwards compatibility)
	})
	for _, major := range majors {
		minors := make([]uint64, 0, len(majorMinorBuckets[major]))
		for minor := range majorMinorBuckets[major] {
			minors = append(minors, minor)
		}
		sort.Slice(minors, func(i, j int) bool {
			return minors[i] > minors[j] // sort descending, minor updates bring both feature and bugfixes
		})
		for _, minor := range minors {
			fmt.Fprintln(os.Stdout)
			fmt.Fprintf(os.Stdout, "Updates to %d.%d:\n", major, minor)
			lastWasLong := false
			headerQueued := true

			// set the minimal cell width to 14 to have a larger space between the columns for shorter versions
			w := tabwriter.NewWriter(os.Stdout, 14, 2, 1, ' ', tabwriter.DiscardEmptyColumns)
			fmt.Fprintf(w, "  VERSION\tISSUES\n")
			// TODO: add metadata about version

			sortConditionalUpdatesBySemanticVersions(majorMinorBuckets[major][minor])
			for i, update := range majorMinorBuckets[major][minor] {
				c := notRecommendedCondition(update)
				if lastWasLong || (c != nil && !showOutdatedReleases) {
					fmt.Fprintln(os.Stdout)
					if c == nil && !headerQueued {
						fmt.Fprintf(w, "  VERSION\tISSUES\n")
						headerQueued = true
					}
					lastWasLong = false
				}
				if i == 2 && !showOutdatedReleases {
					fmt.Fprintf(os.Stdout, "And %d older %d.%d updates you can see with '--show-outdated-releases'.\n", len(majorMinorBuckets[major][minor])-2, major, minor)
					lastWasLong = true
					break
				}
				if c == nil {
					fmt.Fprintf(w, "  %s\t%s\n", update.Release.Version, "no known issues relevant to this cluster")
					if !showOutdatedReleases {
						headerQueued = false
						w.Flush()
					}
				} else if showOutdatedReleases {
					fmt.Fprintf(w, "  %s\t%s\n", update.Release.Version, c.Reason)
				} else {
					fmt.Fprintf(os.Stdout, "  Version: %s\n  Image: %s\n", update.Release.Version, update.Release.Image)
					fmt.Fprintf(os.Stdout, "  Reason: %s\n  Message: %s\n", c.Reason, strings.ReplaceAll(strings.TrimSpace(c.Message), "\n", "\n  "))
					lastWasLong = true
				}
			}
			if showOutdatedReleases {
				w.Flush()
			}
		}
	}
	return nil
}


var UpgradeRecommend = &cobra.Command{
	Use:   "recommend",
	Run: func(cmd *cobra.Command, args []string) {
		admUpgradeRecommendCommand(vars.MustGatherRootPath)
	},
}

func init() {
	UpgradeRecommend.PersistentFlags().BoolVar(&showOutdatedReleases, "show-outdated-releases", false, "")
	UpgradeRecommend.PersistentFlags().BoolVar(&quiet, "quiet", false, "")
}


func notRecommendedCondition(update configv1.ConditionalUpdate) *metav1.Condition {
	if len(update.Risks) == 0 {
		return nil
	}
	if c := findCondition(update.Conditions, "Recommended"); c != nil {
		if c.Status == metav1.ConditionTrue {
			return nil
		}
		return c
	}

	risks := make([]string, len(update.Risks))
	for _, risk := range update.Risks {
		risks = append(risks, risk.Name)
	}
	sort.Strings(risks)
	return &metav1.Condition{
		Type:    "Recommended",
		Status:  "Unknown",
		Reason:  "NoConditions",
		Message: fmt.Sprintf("Conditional update to %s has risks (%s), but no conditions.", update.Release.Version, strings.Join(risks, ", ")),
	}
}



func injectUpgradeableAsCondition(version string, condition *configv1.ClusterOperatorStatusCondition, majorMinorBuckets map[uint64]map[uint64][]configv1.ConditionalUpdate) error {
	current, err := semver.Parse(version)
	if err != nil {
		return fmt.Errorf("cannot parse SemVer version %q: %v", version, err)
	}

	upgradeableURI := fmt.Sprintf("https://docs.openshift.com/container-platform/%d.%d/updating/preparing_for_updates/updating-cluster-prepare.html#cluster-upgradeable_updating-cluster-prepare", current.Major, current.Minor)
	if current.Minor <= 13 {
		upgradeableURI = fmt.Sprintf("https://docs.openshift.com/container-platform/%d.%d/updating/index.html#understanding_clusteroperator_conditiontypes_updating-clusters-overview", current.Major, current.Minor)
	}

	for major, minors := range majorMinorBuckets {
		if major < current.Major {
			continue
		}

		for minor, targets := range minors {
			if major == current.Major && minor <= current.Minor {
				continue
			}

			for i := 0; i < len(targets); i++ {
				majorMinorBuckets[major][minor][i] = ensureUpgradeableRisk(majorMinorBuckets[major][minor][i], condition, upgradeableURI)
			}
		}
	}

	return nil
}


func ensureUpgradeableRisk(target configv1.ConditionalUpdate, condition *configv1.ClusterOperatorStatusCondition, upgradeableURI string) configv1.ConditionalUpdate {
	if hasUpgradeableRisk(target, condition) {
		return target
	}

	target.Risks = append(target.Risks, configv1.ConditionalUpdateRisk{
		URL:           upgradeableURI,
		Name:          "UpgradeableFalse",
		Message:       condition.Message,
		MatchingRules: []configv1.ClusterCondition{{Type: "Always"}},
	})

	for i, c := range target.Conditions {
		if c.Type == "Recommended" {
			if c.Status == metav1.ConditionTrue {
				target.Conditions[i].Reason = condition.Reason
				target.Conditions[i].Message = condition.Message
			} else {
				target.Conditions[i].Reason = "MultipleReasons"
				target.Conditions[i].Message = fmt.Sprintf("%s\n\n%s", condition.Message, c.Message)
			}
			target.Conditions[i].Status = metav1.ConditionFalse
			return target
		}
	}

	target.Conditions = append(target.Conditions, metav1.Condition{
		Type:    "Recommended",
		Status:  metav1.ConditionFalse,
		Reason:  condition.Reason,
		Message: condition.Message,
	})
	return target
}

func hasUpgradeableRisk(target configv1.ConditionalUpdate, condition *configv1.ClusterOperatorStatusCondition) bool {
	for _, risk := range target.Risks {
		if strings.Contains(risk.Message, condition.Message) {
			return true
		}
	}
	return false
}
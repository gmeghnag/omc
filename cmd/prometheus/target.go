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
package prometheus

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

var PrometheusInstance string

var TargetSubCmd = &cobra.Command{
	Use:     "target",
	Aliases: []string{"targets"},
	Short:   "Retrieve the targets (and their status) scraped by Prometheus.",
	Run: func(cmd *cobra.Command, args []string) {
		monitoringExist, _ := helpers.Exists(vars.MustGatherRootPath + "/monitoring")
		if !monitoringExist {
			fmt.Fprintln(os.Stderr, "Path '"+vars.MustGatherRootPath+"/monitoring' does not exist.")
			os.Exit(1)
		}
		alertsFilePath := vars.MustGatherRootPath + "/monitoring/prometheus/" + PrometheusInstance + "/active-targets.json"
		alertsFilePathExist, _ := helpers.Exists(alertsFilePath)
		if !alertsFilePathExist {
			fmt.Fprintln(os.Stderr, "Prometheus targets not found in must-gather.")
			os.Exit(1)
		}
		targets := TargetData{}
		file, _ := os.ReadFile(alertsFilePath)
		err := json.Unmarshal([]byte(file), &targets)
		if err != nil {
			fmt.Println(err)
		}
		headers := []string{"TARGET", "SCRAPE URL", "HEALTH", "LAST ERROR"}
		var data [][]string
		for _, target := range targets.Data.ActiveTargets {
			row := []string{target.DiscoveredLabels["__meta_kubernetes_endpoint_address_target_name"], target.ScrapeURL, target.Health, target.LastError}
			data = append(data, row)
		}
		helpers.PrintTable(headers, data)
	},
}

func init() {
	TargetSubCmd.Flags().StringVarP(&PrometheusInstance, "instance", "i", "prometheus-k8s-0", "Show targets for specific prometheus instance, availables: [prometheus-k8s-0|prometheus-k8s-1].")
}

// Target has the information for one target.
type Target struct {
	// Labels before any processing.
	DiscoveredLabels map[string]string `json:"discoveredLabels"`
	// Any labels that are added to this target and its metrics.
	Labels map[string]string `json:"labels"`

	ScrapePool string `json:"scrapePool"`
	ScrapeURL  string `json:"scrapeUrl"`
	GlobalURL  string `json:"globalUrl"`

	LastError          string              `json:"lastError"`
	LastScrape         time.Time           `json:"lastScrape"`
	LastScrapeDuration float64             `json:"lastScrapeDuration"`
	Health             string              `json:"health"`

	ScrapeInterval string `json:"scrapeInterval"`
	ScrapeTimeout  string `json:"scrapeTimeout"`
}

type DroppedTarget struct {
	// Labels before any processing.
	DiscoveredLabels map[string]string `json:"discoveredLabels"`
}

type TargetDiscovery struct {
	ActiveTargets       []*Target        `json:"activeTargets"`
	DroppedTargets      []*DroppedTarget `json:"droppedTargets"`
	DroppedTargetCounts map[string]int   `json:"droppedTargetCounts"`
}

type TargetData struct {
	Data TargetDiscovery `json:"data"`
}

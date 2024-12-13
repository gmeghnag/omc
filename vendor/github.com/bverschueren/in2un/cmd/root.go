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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	InsightsCmd = &cobra.Command{
		Use:   "in2un",
		Args:  cobra.MinimumNArgs(1),
		Short: "Parse Insights data as unstructed data or raw log lines.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
		},
	}

	ResourceGroup, ResourceName, Namespace, Active, LogLevel string
	AllNamespaces                                            bool
	ConfigDir                                                = "$HOME/.in2un/"
	configFileName                                           = "in2un"
	configFileType                                           = "json"
)

func Execute() {
	err := InsightsCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.AddConfigPath(ConfigDir)

	InsightsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// TODO: fix collision with shorthand "v" set with klog's addGoFlags (omc)
	InsightsCmd.PersistentFlags().StringVar(&LogLevel, "loglevel", "warning", "Logging level")
	InsightsCmd.PersistentFlags().StringVarP(&Namespace, "namespace", "n", "", "If present, the namespace scope for this CLI request")
	InsightsCmd.PersistentFlags().StringVarP(&Active, "insights-file", "", "", "Insights file to read from")

	viper.BindPFlag("Active", InsightsCmd.PersistentFlags().Lookup("Active"))
}

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

	"path/filepath"

	"github.com/bverschueren/in2un/pkg/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var useCmd = &cobra.Command{
	Use:              "use",
	Args:             cobra.MinimumNArgs(1),
	Short:            "Specify the insights file to read from",
	PersistentPreRun: nil,
	Run: func(cmd *cobra.Command, args []string) {
		insightsArchive, _ := filepath.Abs(args[0])
		active, err := reader.NewInsightsReader(insightsArchive)
		if err != nil {
			log.Fatal(err)
		}
		err = os.MkdirAll(ConfigDir, 0750)
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("active", active.Path)
		viper.WriteConfigAs(filepath.Join(ConfigDir, configFileName) + "." + configFileType)
	},
}

func init() {
	InsightsCmd.AddCommand(useCmd)
}

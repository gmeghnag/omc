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
	"io"
	"os"

	"github.com/bverschueren/in2un/pkg/reader"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logsCmd = &cobra.Command{
		Use:   "logs",
		Args:  cobra.MinimumNArgs(1),
		Short: "Return raw log lines from insights data.",
		Run: func(cmd *cobra.Command, args []string) {
			resourceGroup := "pod" // TODO: implement logging for <resource-type>/<resource-name>
			resourceName := args[0]
			ir, err := reader.NewInsightsReader(viper.GetString("active"))
			if err != nil {
				log.Fatal(err)
			}
			found := ir.ReadLog(resourceGroup, resourceName, Namespace, containerName, previous)
			io.Copy(os.Stdout, found)
		},
	}
	containerName string
	previous      bool
)

func init() {
	InsightsCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&containerName, "container", "c", "", "Container to read logs from.")
	logsCmd.Flags().BoolVarP(&previous, "previous", "p", false, "Read from previous logs.")
}

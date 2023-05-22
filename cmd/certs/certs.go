/*
Copyright © 2023 Bram Verschueren <bverschueren@redhat.com>

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
package certs

import (
	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
	"os"
)

var listNonCerts, showParseFailure bool

var Certs = &cobra.Command{
	Use: "certs",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	Certs.AddCommand(
		Inspect,
	)
	Certs.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	Certs.PersistentFlags().BoolVarP(&listNonCerts, "list-non-certs", "", false, "If present, list resources regardless if it contains a certificate.")
	Certs.PersistentFlags().BoolVarP(&showParseFailure, "show-parse-failure", "", false, "If present, list the output of parse attempts for resources.")
	Certs.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide")
}

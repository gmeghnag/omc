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
package machineconfig

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

func checkMachineConfigDiff(first string, second string) {
	firstMachineConfigPath := vars.MustGatherRootPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigs/" + first + ".yaml"
	secondMachineConfigPath := vars.MustGatherRootPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigs/" + second + ".yaml"
	if vars.DiffCmd == "" {
		vars.DiffCmd = "vimdiff"
	}
	_, err := exec.LookPath(vars.DiffCmd)
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(vars.DiffCmd, firstMachineConfigPath, secondMachineConfigPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}

var Diff = &cobra.Command{
	Use:     "diff",
	Aliases: []string{"compare"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "error: two arguments expected, found ", strconv.Itoa(len(args)))
			os.Exit(1)
		}
		checkMachineConfigDiff(args[0], args[1])
	},
}

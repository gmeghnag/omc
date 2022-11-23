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
package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/coreos/go-semver/semver"
	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

var DesiredVersion string

func upgradeBinary(repoName string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	omcExecutablePath := filepath.Dir(ex) + "/omc"
	operatingSystem := runtime.GOOS
	if DesiredVersion == "" {
		checkReleases(repoName)
		os.Exit(0)
	}
	if string(DesiredVersion[0]) != "v" {
		fmt.Println("error: --to must be a semantic version (e.g. v4.0.5): No Major.Minor.Patch elements found")
		os.Exit(1)
	}
	desiredReleaseVer := semver.New(DesiredVersion[1:])
	if vars.OMCVersionTag == "" {
		vars.OMCVersionTag = "v2.0.1"
	}
	currentVer := semver.New(vars.OMCVersionTag[1:])
	if desiredReleaseVer.LessThan(*currentVer) {
		fmt.Println("error: The update " + DesiredVersion + " is not one of the available updates (check them by running \"omc upgrade\")")
		os.Exit(1)
	}
	switch operatingSystem {
	case "windows":
		fmt.Println("This command is not available for windows.")
		fmt.Println("Open an issue on the GitHub repo https://github.com/gmeghnag/omc if you want it impemented.")
	case "darwin":
		err = updateOmcExecutable(omcExecutablePath, "https://github.com/"+repoName+"/releases/download/"+DesiredVersion+"/omc_Darwin_x86_64")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "linux":
		err = updateOmcExecutable(omcExecutablePath, "https://github.com/"+repoName+"/releases/download/"+DesiredVersion+"/omc_Linux_x86_64")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println("This command is not available for the OS you are using.")
		fmt.Println("Open an issue on the GitHub repo https://github.com/gmeghnag/omc if you want it impemented.")
	}
}

// etcdCmd represents the etcd command
var Upgrade = &cobra.Command{
	Use: "upgrade",
	Run: func(cmd *cobra.Command, args []string) {
		upgradeBinary("gmeghnag/omc")
	},
}

func init() {
	Upgrade.Flags().StringVarP(&DesiredVersion, "to", "", "", "Specify the version to upgrade to. The version must be on the list of available updates.")
}

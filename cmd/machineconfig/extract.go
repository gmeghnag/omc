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
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	ign3types "github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/gmeghnag/omc/vars"
	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	ctrlcommon "github.com/openshift/machine-config-operator/pkg/controller/common"
	"github.com/spf13/cobra"
	"github.com/vincent-petithory/dataurl"
	"sigs.k8s.io/yaml"
)

func extractIgnitionConfigStorage(ignConfig ign3types.Config, extractedMachineConfigPath string) {
	ignitionStorageFilesPath := extractedMachineConfigPath + "/storage/files"
	for _, f := range ignConfig.Storage.Files {
		lastInd := strings.LastIndex(f.Path, "/")
		ignitionReferencedFilePath := f.Path[:lastInd]
		_ = os.MkdirAll(ignitionStorageFilesPath+ignitionReferencedFilePath, os.ModePerm)
		if f.Contents.Source != nil {
			contents, err := dataurl.DecodeString(*f.Contents.Source)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			ioutil.WriteFile(ignitionStorageFilesPath+f.Path, contents.Data, 0644)
		}
	}
}

var Extract = &cobra.Command{
	Use: "extract",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "error: one argument expected, found ", strconv.Itoa(len(args)))
			os.Exit(1)
		}
		machineconfigYamlPath := vars.MustGatherRootPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigs/" + args[0] + ".yaml"
		_file, err := ioutil.ReadFile(machineconfigYamlPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		machineConfig := mcfgv1.MachineConfig{}
		if err := yaml.Unmarshal([]byte(_file), &machineConfig); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+machineconfigYamlPath)
			os.Exit(1)
		}
		ignConfig, err := ctrlcommon.ParseAndConvertConfig(machineConfig.Spec.Config.Raw)
		extractedMachineConfigPath := vars.MustGatherRootPath + "/extracted-machine-configs/" + machineConfig.Name
		_ = os.Mkdir(extractedMachineConfigPath, os.ModePerm)
		extractIgnitionConfigStorage(ignConfig, extractedMachineConfigPath)
	},
}

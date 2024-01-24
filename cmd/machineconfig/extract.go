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
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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
			if f.Contents.Compression != nil && *f.Contents.Compression == "gzip" {
				reader := bytes.NewReader([]byte(contents.Data))
				gzreader, e1 := gzip.NewReader(reader)
				if e1 != nil {
					fmt.Fprintln(os.Stderr, e1)
				}

				output, e2 := io.ReadAll(gzreader)
				if e2 != nil {
					fmt.Fprintln(os.Stderr, e2)
				}

				result := string(output)
				fmt.Println(ignitionStorageFilesPath + f.Path)
				os.WriteFile(ignitionStorageFilesPath+f.Path, []byte(result), 0644)
			} else {
				fmt.Println(ignitionStorageFilesPath + f.Path)
				os.WriteFile(ignitionStorageFilesPath+f.Path, contents.Data, 0644)
			}
		}
	}
	ignitionPasswdFilesPath := extractedMachineConfigPath + "/passwd/"
	for _, f := range ignConfig.Passwd.Users {
		_ = os.MkdirAll(ignitionPasswdFilesPath+"users/", os.ModePerm)
		keys := ""
		for _, key := range f.SSHAuthorizedKeys {
			keys = keys + string(key) + "\n"
		}
		os.WriteFile(ignitionPasswdFilesPath+"users/"+f.Name, []byte(keys), 0644)
		fmt.Println(ignitionPasswdFilesPath + "users/" + f.Name)
	}
	ignitionSystemdPath := extractedMachineConfigPath + "/systemd/units/"
	for _, f := range ignConfig.Systemd.Units {
		if len(f.Dropins) > 0 {
			_ = os.MkdirAll(ignitionSystemdPath+f.Name+"/dropins", os.ModePerm)
			for _, dropin := range f.Dropins {
				var content []byte
				if dropin.Contents != nil {
					content = []byte(*dropin.Contents)
				}
				os.WriteFile(ignitionSystemdPath+f.Name+"/dropins/"+dropin.Name, content, 0644)
				fmt.Println(ignitionSystemdPath + f.Name + "/dropins/" + dropin.Name)
			}
		} else {
			fmt.Println(ignitionSystemdPath + f.Name)
			var content []byte
			if f.Contents != nil {
				content = []byte(*f.Contents)
			}
			os.WriteFile(ignitionSystemdPath+f.Name, content, 0644)
		}

		//keys := ""
		//for _, key := range f.SSHAuthorizedKeys {
		//	keys = keys + string(key) + "\n"
		//}
		//os.WriteFile(ignitionPasswdFilesPath+"users/"+f.Name, []byte(keys), 0644)
		//fmt.Println(ignitionPasswdFilesPath + "users/" + f.Name)
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
		_file, err := os.ReadFile(machineconfigYamlPath)
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

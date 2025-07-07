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

func extractMachineConfig(machineConfigName string) error {
	machineconfigYamlPath := vars.MustGatherRootPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigs/" + machineConfigName + ".yaml"
	_file, err := os.ReadFile(machineconfigYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read machine config %s: %w", machineConfigName, err)
	}

	machineConfig := mcfgv1.MachineConfig{}
	if err := yaml.Unmarshal([]byte(_file), &machineConfig); err != nil {
		return fmt.Errorf("failed to unmarshal machine config %s: %w", machineConfigName, err)
	}

	ignConfig, err := ctrlcommon.ParseAndConvertConfig(machineConfig.Spec.Config.Raw)
	if err != nil {
		return fmt.Errorf("failed to parse ignition config for %s: %w", machineConfigName, err)
	}

	extractedMachineConfigPath := vars.MustGatherRootPath + "/extracted-machine-configs/" + machineConfig.Name
	_ = os.MkdirAll(extractedMachineConfigPath, os.ModePerm)

	fmt.Printf("Extracting machine config: %s\n", machineConfig.Name)
	extractIgnitionConfigStorage(ignConfig, extractedMachineConfigPath)

	return nil
}

var extractAll bool

var Extract = &cobra.Command{
	Use:   "extract [machine-config-name]",
	Short: "Extract files from a MachineConfig",
	Long: `Extract files from a MachineConfig resource and save them to the local filesystem.

The command requires either the name of a specific MachineConfig as an argument, or the --all flag to extract all MachineConfigs.

Examples:
  omc machine-config extract 00-master
  omc machine-config extract 00-worker
  omc machine-config extract rendered-master-1234567890
  omc machine-config extract --all

To list available MachineConfigs, use:
  omc get machineconfigs`,
	Run: func(cmd *cobra.Command, args []string) {
		if extractAll {
			if len(args) > 0 {
				fmt.Fprintln(os.Stderr, "error: cannot specify machine config name when using --all flag")
				os.Exit(1)
			}

			// Extract all machine configs
			machineConfigsDir := vars.MustGatherRootPath + "/cluster-scoped-resources/machineconfiguration.openshift.io/machineconfigs/"
			files, err := os.ReadDir(machineConfigsDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: failed to read machine configs directory: %v\n", err)
				os.Exit(1)
			}

			var machineConfigNames []string
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
					// Remove .yaml extension to get the machine config name
					name := strings.TrimSuffix(file.Name(), ".yaml")
					machineConfigNames = append(machineConfigNames, name)
				}
			}

			if len(machineConfigNames) == 0 {
				fmt.Fprintln(os.Stderr, "error: no machine configs found")
				os.Exit(1)
			}

			fmt.Printf("Found %d machine config(s) to extract\n", len(machineConfigNames))

			for _, name := range machineConfigNames {
				if err := extractMachineConfig(name); err != nil {
					fmt.Fprintf(os.Stderr, "error extracting %s: %v\n", name, err)
				}
			}

			fmt.Printf("Extraction complete. Files saved to: %s/extracted-machine-configs/\n", vars.MustGatherRootPath)
		} else {
			// Extract single machine config
			if len(args) != 1 {
				fmt.Fprintln(os.Stderr, "error: one argument expected, found ", strconv.Itoa(len(args)))
				os.Exit(1)
			}

			if err := extractMachineConfig(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	Extract.Flags().BoolVar(&extractAll, "all", false, "Extract all available machine configs")
}

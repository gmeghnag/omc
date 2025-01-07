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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// follow kubectl logic and expect args to be either:
// a resource as single argument: "<recource-type>"
// a resource and its name as a single argument seperated by a slash: "<recource-type>/<resource-name>"
// a resource and its name as sequential arguments: "<recource-type>/<resource-name>"
func processArgs(args []string) (string, string) {
	var resourceGroup, resourceName string
	if len(args) == 1 && strings.Contains(args[0], "/") {
		parts := strings.Split(args[0], "/")
		resourceGroup = parts[0]
		resourceName = parts[1]
	} else {
		resourceGroup = args[0]
		if len(args) > 1 {
			resourceName = args[1]
		}
	}
	resourceGroup = Unalias(resourceGroup)
	return resourceGroup, resourceName
}

func Unalias(alias string) string {
	log.Debug("Using static alias map as best effort")
	aliases := map[string]string{
		"mc":             "machineconfig",
		"mcp":            "machineconfigpool",
		"cm":             "configmap",
		"co":             "clusteroperator",
		"ns":             "namespace",
		"clusterversion": "version",
	}
	if unalias, ok := aliases[alias]; ok {
		return unalias
	}
	return alias
}

func initConfig() {
	level, err := log.ParseLevel(LogLevel)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	log.SetLevel(level)

	ConfigDir = os.ExpandEnv(ConfigDir)

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)

	if err := viper.ReadInConfig(); err != nil {
		configFile := filepath.Join(ConfigDir, configFileName) + "." + configFileType
		if os.IsNotExist(err) {
			log.Debugf("Writing config to %s\n", configFile)
			viper.WriteConfigAs(configFile)
		} else {
			// ReadInConfig returns a viper-defined error on not found
			// https://github.com/spf13/viper/blob/54f2089833b65fa556a510957197de2609059147/viper.go#L1483
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Debugf("Writing config to %s\n", configFile)
				viper.WriteConfigAs(configFile)
			} else {
				fmt.Println("Can't read config:", err)
			}
		}
	} else {
		log.Debugf("Active insights archive: %s\n", viper.GetString("Active"))
	}
}

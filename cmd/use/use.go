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
package use

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

func useContext(path string, omcConfigFile string, idFlag string) {
	if path != "" {
		_path, err := findMustGatherIn(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		l := strings.Split(_path, "/")
		path = strings.Join(l[0:(len(l)-1)], "/")
		path = strings.TrimSuffix(path, "/")
	}

	// read json omcConfigFile
	file, _ := os.ReadFile(omcConfigFile)
	omcConfigJson := types.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)

	config := types.Config{}
	var contexts []types.Context
	var NewContexts []types.Context
	contexts = omcConfigJson.Contexts
	var found bool
	var ctxId string
	for _, c := range contexts {
		if c.Id == idFlag || c.Path == path {
			NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: "*", Project: c.Project})
			found = true
		} else {
			NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: "", Project: c.Project})
		}
	}
	if !found {
		if idFlag != "" {
			NewContexts = append(NewContexts, types.Context{Id: idFlag, Path: path, Current: "*", Project: "default"})
		} else {
			ctxId = helpers.RandString(8)
			var namespaces []string
			_namespaces, _ := os.ReadDir(path + "/namespaces/")
			for _, f := range _namespaces {
				namespaces = append(namespaces, f.Name())
			}
			if len(namespaces) == 1 {
				NewContexts = append(NewContexts, types.Context{Id: ctxId, Path: path, Current: "*", Project: namespaces[0]})
			} else if len(namespaces) > 1 && slices.Contains(namespaces, "openshift-logging") {
				NewContexts = append(NewContexts, types.Context{Id: ctxId, Path: path, Current: "*", Project: "openshift-logging"})
			} else {
				NewContexts = append(NewContexts, types.Context{Id: ctxId, Path: path, Current: "*", Project: "default"})
			}
		}

	}

	config.Contexts = NewContexts
	config.Id = idFlag
	if !found {
		if idFlag != "" {
			config.Id = idFlag
		} else {
			config.Id = ctxId
		}
	}
	file, _ = json.MarshalIndent(config, "", " ")
	_ = os.WriteFile(omcConfigFile, file, 0644)

}

func findMustGatherIn(path string) (string, error) {
	numDirs := 0
	dirName := ""
	retPath := strings.TrimSuffix(path, "/")
	var retErr error
	timeStampFound := false
	resourcesFolderFound := false
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			dirName = file.Name()
			numDirs = numDirs + 1
			if file.Name() == "namespaces" || file.Name() == "cluster-scoped-resources" {
				resourcesFolderFound = true
			}
		}
		if !file.IsDir() && file.Name() == "timestamp" {
			timeStampFound = true
		}
	}
	if numDirs == 1 && !timeStampFound && !resourcesFolderFound {
		return findMustGatherIn(path + "/" + dirName)
	}
	if resourcesFolderFound {
		return retPath + "/", retErr
	}
	if timeStampFound && (numDirs > 1 || numDirs == 0) {
		return path, fmt.Errorf("expected one directory in path: \"%s\", found: %s", path, strconv.Itoa(numDirs))
	}
	if !timeStampFound && !resourcesFolderFound {
		// Case: "path" is an empty directory
		return path, fmt.Errorf("wrong must-gather file composition for %v", path)
	}
	return findMustGatherIn(path + "/" + dirName)
}

// useCmd represents the use command
var UseCmd = &cobra.Command{
	Use:   "use",
	Short: "Select the must-gather to use",
	Long: `
	Select the must-gather to use.
	If the must-gather does not exists it will be added as default to the managed must-gathers.
	Use the command 'omc get mg' to see them all.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		idFlag, _ := cmd.Flags().GetString("id")
		path := ""
		isCompressedFile := false
		if len(args) == 0 && idFlag == "" {
			fmt.Printf("Must-Gather  : %s\nProject      : %s\n", vars.MustGatherRootPath, vars.Namespace)
			InfrastrctureFilePathExists, _ := helpers.Exists(vars.MustGatherRootPath + "/cluster-scoped-resources/config.openshift.io/infrastructures.yaml")
			if InfrastrctureFilePathExists {
				_file, _ := os.ReadFile(vars.MustGatherRootPath + "/cluster-scoped-resources/config.openshift.io/infrastructures.yaml")
				infrastructureList := configv1.InfrastructureList{}
				if err := yaml.Unmarshal([]byte(_file), &infrastructureList); err != nil {
					fmt.Println("Error when trying to unmarshal file: " + vars.MustGatherRootPath + "/cluster-scoped-resources/config.openshift.io/infrastructures.yaml")
					os.Exit(1)
				} else {
					fmt.Printf("ApiServerURL : %s\n", infrastructureList.Items[0].Status.APIServerURL)
					fmt.Printf("Platform     : %s\n", infrastructureList.Items[0].Status.PlatformStatus.Type)
				}
			}
			os.Exit(0)
		}
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Expect one argument, found: ", len(args))
			os.Exit(1)
		}
		if len(args) == 1 {
			path = args[0]
			if IsRemoteFile(path) {
				path, err = DownloadFile(path)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			} else {
				if strings.HasSuffix(path, "/") {
					path = strings.TrimRight(path, "/")
				}
				if strings.HasSuffix(path, "\\") {
					path = strings.TrimRight(path, "\\")
				}
				path, _ = filepath.Abs(path)
			}

			isDir, _ := helpers.IsDirectory(path)
			if !isDir {
				isCompressedFile, _ = IsCompressedFile(path)
				if !isCompressedFile {
					fmt.Fprintln(os.Stderr, "Error: "+path+" is not a directory not a compressed file.")
					os.Exit(1)
				}
			}
		}

		if isCompressedFile {
			outputpath := filepath.Dir(path)
			rootfile, err := DecompressFile(path, outputpath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: decompressing "+path+" in "+outputpath+": "+err.Error())
				os.Exit(1)
			}
			path = rootfile
		}

		useContext(path, viper.ConfigFileUsed(), idFlag)
	},
}

func init() {
	UseCmd.Flags().StringVarP(&vars.Id, "id", "i", "", "Id string for the must-gather to use. If two must-gather has the same id the first one will be used.")
}

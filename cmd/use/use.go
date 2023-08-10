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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	file, _ := ioutil.ReadFile(omcConfigFile)
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
			NewContexts = append(NewContexts, types.Context{Id: ctxId, Path: path, Current: "*", Project: "default"})
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
	_ = ioutil.WriteFile(omcConfigFile, file, 0644)

}

func findMustGatherIn(path string) (string, error) {
	numDirs := 0
	dirName := ""
	retPath := strings.TrimSuffix(path, "/")
	var retErr error
	timeStampFound := false
	namespacesFolderFound := false
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			dirName = file.Name()
			numDirs = numDirs + 1
			if file.Name() == "namespaces" {
				namespacesFolderFound = true
			}
		}
		if !file.IsDir() && file.Name() == "timestamp" {
			timeStampFound = true
		}
	}
	if namespacesFolderFound {
		return retPath + "/", retErr
	}
	if timeStampFound && (numDirs > 1 || numDirs == 0) {
		return path, fmt.Errorf("Expected one directory in path: \"%s\", found: %s.", path, strconv.Itoa(numDirs))
	}
	if !timeStampFound && numDirs == 1 {
		return findMustGatherIn(path + "/" + dirName)
	}
	if !timeStampFound && !namespacesFolderFound {
		// Case: "path" is an empty directory
		return path, fmt.Errorf("wrong must-gather file composition for %v", path)
	}

	return retPath + "/", retErr
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
		idFlag, _ := cmd.Flags().GetString("id")
		path := ""
		isCompressedFile := false
		if len(args) == 0 && idFlag == "" {
			fmt.Printf("must-gather: \"%s\"\nnamespace: \"%s\"\n", vars.MustGatherRootPath, vars.Namespace)
			os.Exit(0)
		}
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Expect one argument, found: ", len(args))
			os.Exit(1)
		}
		if len(args) == 1 {
			path = args[0]
			if strings.HasSuffix(path, "/") {
				path = strings.TrimRight(path, "/")
			}
			if strings.HasSuffix(path, "\\") {
				path = strings.TrimRight(path, "\\")
			}
			path, _ = filepath.Abs(path)
			isDir, _ := helpers.IsDirectory(path)
			if !isDir {
				isCompressedFile, _ = IsCompressedFile(path)
				if (!isCompressedFile) {
				     fmt.Fprintln(os.Stderr, "Error: "+path+" is not a directory not a compressed file.")
					 os.Exit(1)
				}
			}
		}

		if (isCompressedFile) {
			outputpath := filepath.Dir(path)
			rootfile,err := DecompressFile(path,outputpath)
			if ( err != nil ) {
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

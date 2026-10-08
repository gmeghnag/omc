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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/pkg/mustgather"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

var singleNamespaceInMustGather bool

// findExistingContextByRootDir checks if a root directory is already in the
// contexts. It matches on a full path component so an archive root like
// "must-gather" cannot be satisfied by an unrelated path that merely contains
// that substring (e.g. "old-must-gather-backup").
func findExistingContextByRootDir(rootDir string, contexts []types.Context) (string, bool) {
	for _, ctx := range contexts {
		for _, part := range strings.Split(filepath.ToSlash(ctx.Path), "/") {
			if part == rootDir {
				return ctx.Path, true
			}
		}
	}
	return "", false
}

func useContext(path string, omcConfigFile string, idFlag string, rootPaths []string) error {
	// When more than one must-gather was discovered under the used directory the
	// context groups them: rootPaths are already resolved roots ordered
	// most-recent-first, so the normalization below (which unwraps a single
	// must-gather) must be skipped.
	multi := len(rootPaths) > 1
	if path != "" {
		if multi {
			path = strings.TrimSuffix(rootPaths[0], "/")
			vars.MustGatherRootPath = path
			vars.MustGatherRootPaths = rootPaths
		} else {
			_path, err := findMustGatherIn(path)
			if err != nil {
				return err
			}
			l := strings.Split(_path, "/")
			path = strings.Join(l[0:(len(l)-1)], "/")
			path = strings.TrimSuffix(path, "/")
			vars.MustGatherRootPath = path
			vars.MustGatherRootPaths = []string{path}
		}
	}
	// persistPaths is only stored for a grouped context; a single must-gather
	// context keeps its original JSON shape (no "paths" key).
	var persistPaths []string
	if multi {
		persistPaths = rootPaths
	}

	// read json omcConfigFile
	file, _ := os.ReadFile(omcConfigFile)
	omcConfigJson := types.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)

	var contexts []types.Context
	var NewContexts []types.Context
	contexts = omcConfigJson.Contexts
	var found bool
	var ctxId, configId string
	defaultProject := omcConfigJson.DefaultProject
	if defaultProject == "" {
		defaultProject = "default"
	}
	for _, c := range contexts {
		if c.Id == idFlag || c.Path == path {
			// When this invocation resolved a path, the grouping reflects the
			// current discovery (persistPaths: the roots for a multi
			// must-gather, nil to clear a previous grouping for a single one).
			// Selecting an existing context by id alone (no path) keeps its
			// stored grouping.
			ctxPaths := c.Paths
			if path != "" {
				ctxPaths = persistPaths
			}
			NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: "*", Project: c.Project, Paths: ctxPaths})
			configId = c.Id
			found = true
			vars.Namespace = c.Project
		} else {
			// Preserve any other context's grouped roots untouched.
			NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: "", Project: c.Project, Paths: c.Paths})
		}
	}
	if !found {
		if idFlag != "" {
			NewContexts = append(NewContexts, types.Context{Id: idFlag, Path: path, Current: "*", Project: defaultProject, Paths: persistPaths})
		} else {
			ctxId = helpers.RandString(8)
			var namespaces []string
			_namespaces, _ := os.ReadDir(path + "/namespaces/")
			for _, f := range _namespaces {
				namespaces = append(namespaces, f.Name())
			}
			if len(namespaces) == 1 {
				NewContexts = append(NewContexts, types.Context{Id: ctxId, Path: path, Current: "*", Project: namespaces[0], Paths: persistPaths})
				vars.Namespace = namespaces[0]
				singleNamespaceInMustGather = true
			} else {
				NewContexts = append(NewContexts, types.Context{Id: ctxId, Path: path, Current: "*", Project: defaultProject, Paths: persistPaths})
				vars.Namespace = defaultProject
			}
		}

	}

	if !found {
		if idFlag != "" {
			configId = idFlag
		} else {
			configId = ctxId
		}
	}
	config := types.Config{
		Id:             configId,
		Contexts:       NewContexts,
		DiffCmd:        omcConfigJson.DiffCmd,
		DefaultProject: omcConfigJson.DefaultProject,
	}
	file, _ = json.MarshalIndent(config, "", " ")
	_ = os.WriteFile(omcConfigFile, file, 0644)
	return nil
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

func MustGatherInfo() {
	roots := vars.MustGatherRootPaths
	if len(roots) == 0 {
		roots = []string{vars.MustGatherRootPath}
	}

	if len(roots) > 1 {
		fmt.Printf("Must-Gathers   : %d (merged, most-recent-first)\n", len(roots))
		for i, p := range roots {
			marker := "  "
			if i == 0 {
				marker = "* " // the most recent root backs describe/logs/info
			}
			fmt.Printf("  %s%s\n", marker, p)
		}
	} else {
		fmt.Printf("Must-Gather    : %s\n", vars.MustGatherRootPath)
	}
	if singleNamespaceInMustGather {
		fmt.Printf("Project        : %s (single project)\n", vars.Namespace)
	} else {
		fmt.Printf("Project        : %s\n", vars.Namespace)
	}

	// Each field is resolved from the most recent root that actually has it, so
	// a grouped context shows a complete summary even when the most recent
	// capture (e.g. an `oc adm inspect`) is missing some files.

	const infraRel = "cluster-scoped-resources/config.openshift.io/infrastructures.yaml"
	if root, ok := mustgather.ResolveRoot(roots, infraRel); ok {
		_file, _ := os.ReadFile(root + "/" + infraRel)
		infrastructureList := configv1.InfrastructureList{}
		if err := yaml.Unmarshal([]byte(_file), &infrastructureList); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + root + "/" + infraRel)
			os.Exit(1)
		} else if len(infrastructureList.Items) > 0 {
			status := infrastructureList.Items[0].Status
			fmt.Printf("ApiServerURL   : %s\n", status.APIServerURL)
			if status.PlatformStatus != nil {
				fmt.Printf("Platform       : %s\n", status.PlatformStatus.Type)
			}
		}
	}

	const cvRel = "cluster-scoped-resources/config.openshift.io/clusterversions/version.yaml"
	if root, ok := mustgather.ResolveRoot(roots, cvRel); ok {
		_file, _ := os.ReadFile(root + "/" + cvRel)
		ClusterVersion := configv1.ClusterVersion{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterVersion); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + root + "/" + cvRel)
			os.Exit(1)
		} else {
			clusterversion := ""
			for _, version := range ClusterVersion.Status.History {
				if version.State == "Completed" {
					clusterversion = version.Version
					break
				}
			}
			fmt.Printf("ClusterID      : %s\n", ClusterVersion.Spec.ClusterID)
			fmt.Printf("ClusterVersion : %s\n", clusterversion)
		}
	}

	// must-gather.logs sits next to a root; use the first capture that has it.
	for _, root := range roots {
		if clientVersion := extractClientVersion(filepath.Dir(root) + "/must-gather.logs"); clientVersion != "" {
			fmt.Printf("ClientVersion  : %s\n", clientVersion)
			break
		}
	}

	// The image digest is embedded in a must-gather's directory name (an
	// inspect root has none), so use the first capture whose name carries it.
	for _, root := range roots {
		last := filepath.Base(root)
		if strings.Contains(last, "-sha256") {
			fmt.Printf("Image          : %s\n", strings.Split(last, "-sha256")[0])
			break
		}
	}

	// The timestamp file lives either next to the root or inside it (inspect
	// writes it inside). Use the first capture that has one.
	timestamp := "MISSING"
	for _, root := range roots {
		if ts, ok := timestampFromFile(filepath.Join(root, "..", "timestamp")); ok {
			timestamp = ts
			break
		}
		if ts, ok := timestampFromFile(filepath.Join(root, "timestamp")); ok {
			timestamp = ts
			break
		}
	}
	fmt.Println("Timestamp      : " + timestamp)
}

// timestampFromFile parses a must-gather timestamp file, returning the
// "start - end" range (or "INCOMPLETE" when it holds fewer than two lines) and
// whether the file existed.
func timestampFromFile(path string) (string, bool) {
	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var times []string
	for scanner.Scan() {
		rLine := scanner.Text()
		var tLine string
		if i := strings.Index(rLine, " m="); i >= 0 {
			tLine = rLine[:i]
		}
		t, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tLine)
		times = append(times, t.Format("2006-01-02 15:04:05"))
	}
	if len(times) > 1 {
		return times[0] + " - " + times[1], true
	}
	return "INCOMPLETE", true
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
		fileType := ""
		isCompressedFile := false
		if len(args) == 0 && idFlag == "" {
			MustGatherInfo()
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
				isCompressedFile, fileType, _ = IsCompressedFile(path)
				if !isCompressedFile {
					fmt.Fprintln(os.Stderr, "Error: "+path+" is not a directory not a compressed file.")
					os.Exit(1)
				}
			}
		}

		if isCompressedFile {
			outputpath := filepath.Dir(path)

			// Peek into the archive to determine what directory it will create
			rootDirName, err := GetArchiveRootDir(path, fileType)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: unable to determine archive structure: "+err.Error())
				os.Exit(1)
			}

			expectedExtractedPath := filepath.Join(outputpath, rootDirName)

			// Check if the directory already exists
			if info, err := os.Stat(expectedExtractedPath); err == nil && info.IsDir() {
				// Verify it looks like a must-gather directory
				namespacesPath := filepath.Join(expectedExtractedPath, "namespaces")
				clusterPath := filepath.Join(expectedExtractedPath, "cluster-scoped-resources")
				timestampPath := filepath.Join(expectedExtractedPath, "timestamp")

				isMustGather := false
				if _, err := os.Stat(namespacesPath); err == nil {
					isMustGather = true
				} else if _, err := os.Stat(clusterPath); err == nil {
					isMustGather = true
				} else if _, err := os.Stat(timestampPath); err == nil {
					isMustGather = true
				}

				if isMustGather {
					fmt.Println("Archive already extracted at: " + expectedExtractedPath)
					path = expectedExtractedPath
					isCompressedFile = false
				} else {
					fmt.Println("Directory exists but doesn't appear to be a must-gather, extracting archive...")
				}
			} else {
				// Directory doesn't exist, check if it's in omc.json (maybe extracted elsewhere)
				file, _ := os.ReadFile(viper.ConfigFileUsed())
				omcConfigJson := types.Config{}
				_ = json.Unmarshal([]byte(file), &omcConfigJson)

				if existingPath, found := findExistingContextByRootDir(rootDirName, omcConfigJson.Contexts); found {
					// Verify the path still exists
					if _, err := os.Stat(existingPath); err == nil {
						fmt.Println("Archive already registered in omc.json, using: " + existingPath)
						path = existingPath
						isCompressedFile = false
					} else {
						fmt.Println("Previous extraction no longer exists at " + existingPath + ", extracting archive...")
					}
				}
			}

			// If still marked as compressed, extract it
			if isCompressedFile {
				rootfile, err := DecompressFile(path, outputpath, fileType)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error: decompressing "+path+" in "+outputpath+": "+err.Error())
					os.Exit(1)
				}
				path = rootfile
			}
		}

		// Auto-discover every must-gather under the used directory. More than
		// one turns this into a grouped context that get/describe/logs/events
		// read across; a single one keeps the original single-must-gather flow.
		var rootPaths []string
		if roots, derr := mustgather.DiscoverRoots(path); derr == nil && len(roots) > 1 {
			rootPaths = mustgather.Paths(roots)
		}
		err = useContext(path, viper.ConfigFileUsed(), idFlag, rootPaths)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		MustGatherInfo()
	},
}

func init() {
	UseCmd.Flags().StringVarP(&vars.Id, "id", "i", "", "Id string for the must-gather to use. If two must-gather has the same id the first one will be used.")
}

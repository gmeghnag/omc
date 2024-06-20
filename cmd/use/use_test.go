package use

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"
)

func TestUseContext(t *testing.T) {
	singleNSPath := "testdata/useContext/must-gather-singleNS.sample"
	multipleNSPath := "testdata/useContext/must-gather-multipleNS.sample"
	testCases := []struct {
		name           string
		path           string
		ignoreConfigID bool // ID will be randomly created so we sometimes need to ignore the generated ID
		idFlag         string
		initialFile    string // JSON content for the omcConfigFile
		expectedFile   string // Expected JSON content after the function call
	}{
		{
			name: "test1",
			path: singleNSPath,
			initialFile: `{
                "id": "eFqOopAs",
				"contexts": [
					{
					 "id": "eFqOopAs",
					 "path": "testdata/useContext/must-gather-singleNS.sample",
					 "current": "*",
					 "project": "openshift-etcd"
					}
				   ],
                "default_project": "openshift-etcd"
            }`,
			expectedFile: `{
                "id": "eFqOopAs",
				"contexts": [
					{
					 "id": "eFqOopAs",
					 "path": "testdata/useContext/must-gather-singleNS.sample",
					 "current": "*",
					 "project": "openshift-etcd"
					}
				   ],
                "default_project": "openshift-etcd"
            }`,
		},
		{
			name: "test2",
			path: multipleNSPath,
			initialFile: `{
                "id": "eFqOopAs",
				"contexts": [
					{
					 "id": "eFqOopAs",
					 "path": "testdata/useContext/must-gather-multipleNS.sample",
					 "current": "*",
					 "project": "openshift-etcd"
					}
				   ],
                "default_project": "openshift-etcd"
            }`,
			expectedFile: `{
                "id": "eFqOopAs",
				"contexts": [
					{
					 "id": "eFqOopAs",
					 "path": "testdata/useContext/must-gather-multipleNS.sample",
					 "current": "*",
					 "project": "openshift-etcd"
					}
				   ],
                "default_project": "openshift-etcd"
            }`,
		},
		{
			name:           "test3",
			path:           singleNSPath,
			ignoreConfigID: true,
			initialFile: `{
                "default_project": "openshift-etcd"
            }`,
			expectedFile: `{
				"id": "WuLN4pDY",
				"contexts": [
				 {
				  "id": "WuLN4pDY",
				  "path": "testdata/useContext/must-gather-singleNS.sample",
				  "current": "*",
				  "project": "openshift-etcd"
				 }
				],
				"default_project": "openshift-etcd"
            }`,
		},
		{
			name:           "test4",
			path:           multipleNSPath,
			ignoreConfigID: true,
			initialFile: `{
            }`,
			expectedFile: `{
				"id": "WuLN4pDY",
				"contexts": [
				 {
				  "id": "WuLN4pDY",
				  "path": "testdata/useContext/must-gather-multipleNS.sample",
				  "current": "*",
				  "project": "default"
				 }
				]
            }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup temporary directory for testing
			tempDir, err := ioutil.TempDir("", "testusecontext")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir) // Clean up

			// Prepare the omcConfigFile path and initial content
			configFilePath := filepath.Join(tempDir, "omc.json")
			if err := ioutil.WriteFile(configFilePath, []byte(tc.initialFile), 0666); err != nil {
				t.Fatalf("Unable to write initial file: %v", err)
			}

			useContext(tc.path, configFilePath, tc.idFlag)

			// Validate the file after function call
			writtenFileContent, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				t.Fatalf("Failed to read back the file: %v", err)
			}
			// Normalize JSON contents for comparison, since JSON formatting may differ
			var actualConfig, expectedConfig types.Config
			if err := json.Unmarshal(writtenFileContent, &actualConfig); err != nil {
				t.Fatalf("Failed to unmarshal actual file: %v", err)
			}
			if err := json.Unmarshal([]byte(tc.expectedFile), &expectedConfig); err != nil {
				t.Fatalf("Failed to unmarshal expected file: %v", err)
			}

			// Compare actual config with expected config
			if !compareConfig(actualConfig, expectedConfig, tc.ignoreConfigID) {
				t.Errorf("Expected file content %+v, got %+v", expectedConfig, actualConfig)
			}

			// Ensure state matches the current path and project
			checkContextConsistency(t, actualConfig)
		})
	}
}

func compareContexts(actualContext []types.Context, expectedContext []types.Context, ignoreContextID bool) bool {
	if len(actualContext) != len(expectedContext) {
		return false
	}

	for i, actual := range actualContext {
		expected := expectedContext[i]

		if !ignoreContextID && actual.Id == "" {
			if actual.Id != expected.Id {
				return false
			}
		}
		if actual.Path != expected.Path {
			return false
		}
		if actual.Project != expected.Project {
			return false
		}
	}

	return true
}

func compareConfig(actualConfig types.Config, expectedConfig types.Config, ignoreConfigID bool) bool {

	if !ignoreConfigID && actualConfig.Id != expectedConfig.Id {
		return false
	}

	if actualConfig.DefaultProject != expectedConfig.DefaultProject ||
		actualConfig.DiffCmd != expectedConfig.DiffCmd ||
		actualConfig.UseLocalCRDs != expectedConfig.UseLocalCRDs {
		return false
	}

	// This line executes after all the condition checks.
	return compareContexts(actualConfig.Contexts, expectedConfig.Contexts, ignoreConfigID)
}

// check that what's stored in memory matches the currently selected context
func checkContextConsistency(t *testing.T, config types.Config) {
	t.Helper()
	c, err := currentContext(config.Contexts)
	if err != nil {
		t.Error(err)
	}

	if c.Path != vars.MustGatherRootPath {
		t.Errorf("in-memory path does not match current context. Want %q, got %q",
			c.Path,
			vars.MustGatherRootPath,
		)
	}

	if c.Project != vars.Namespace {
		t.Errorf("in-memory project does not match current context. Want %q, got %q",
			c.Project,
			vars.Namespace,
		)
	}
}

func currentContext(contexts []types.Context) (types.Context, error) {
	for _, c := range contexts {
		if c.Current == "*" {
			return c, nil
		}
	}

	return types.Context{}, errors.New("could not find current context")
}

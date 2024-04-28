package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"
)

// Define a struct to hold each test case
type testCase struct {
	name           string
	useLocalCRDs   bool
	diffCmd        string
	defaultProject string
}

func TestSetConfig(t *testing.T) {
	// Define your test cases
	tests := []testCase{
		{
			name:           "Test Case 1",
			useLocalCRDs:   true,
			diffCmd:        "diff",
			defaultProject: "project1",
		},
		{
			name:           "Test Case 2",
			useLocalCRDs:   false,
			diffCmd:        "diff --unified",
			defaultProject: "project2",
		},
		// Add more tests as needed
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := ioutil.TempDir("", "omcConfigTest")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create '.omc' directory inside the temp directory
			omcDir := filepath.Join(tempDir, ".omc")
			if err := os.MkdirAll(omcDir, 0755); err != nil {
				t.Fatalf("Failed to create '.omc' directory: %v", err)
			}

			// Set test variables according to the current test case
			vars.UseLocalCRDs = tc.useLocalCRDs
			vars.DiffCmd = tc.diffCmd
			vars.DefaultProject = tc.defaultProject

			// Temporarily set the HOME environment variable to the temp directory
			originalHome := os.Getenv("HOME")
			if err := os.Setenv("HOME", tempDir); err != nil {
				t.Fatalf("Failed to set temporary HOME environment variable: %v", err)
			}
			defer os.Setenv("HOME", originalHome) // Restore the original HOME value after the test

			// Invoke the function under test
			SetConfig()

			// Read the config file from the temp dir '.omc' directory
			testConfigPath := filepath.Join(omcDir, "omc.json")
			file, err := ioutil.ReadFile(testConfigPath)
			if err != nil {
				t.Fatalf("Unable to read config file from temp dir: %v", err)
			}

			var config types.Config
			err = json.Unmarshal(file, &config)
			if err != nil {
				t.Fatalf("Unable to unmarshal config JSON: %v", err)
			}
			if !compareConfig(config, tc) {
				t.Errorf("Expected config %+v, got %+v", tc, config)
			}
		})
	}
}

func compareConfig(actualConfig types.Config, tc testCase) bool {
	if actualConfig.UseLocalCRDs != tc.useLocalCRDs {
		return false
	}
	if actualConfig.DiffCmd != tc.diffCmd {
		return false
	}
	if actualConfig.DefaultProject != tc.defaultProject {
		return false
	}
	return true
}

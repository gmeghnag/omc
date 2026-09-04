// Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

package configpath

import (
	"os"
	"path/filepath"
)

const (
	EnvConfigHome        = "OMC_CONFIG_HOME"
	DefaultConfigDirName = ".omc"
	ConfigFileName       = "omc.json"
	CRDsDirName          = "customresourcedefinitions"
	PullSecretTxtName    = "pull-secret.txt"
	PullSecretJsonName   = "pull-secret.json"
)

// Resolver holds the configuration directory resolution logic
type Resolver struct {
	configDir string
}

// NewResolver creates a resolver with precedence: flag > env > default
func NewResolver(flagValue string) *Resolver {
	var configDir string

	if flagValue != "" {
		// CLI flag takes precedence
		configDir = flagValue
	} else if envValue := os.Getenv(EnvConfigHome); envValue != "" {
		// Environment variable is second priority
		configDir = envValue
	} else {
		// Default to ~/.omc
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home cannot be determined
			configDir = DefaultConfigDirName
		} else {
			configDir = filepath.Join(home, DefaultConfigDirName)
		}
	}

	// Ensure we have an absolute path
	absPath, err := filepath.Abs(configDir)
	if err != nil {
		// If we can't get absolute path, use as-is
		absPath = configDir
	}

	return &Resolver{configDir: absPath}
}

// GetConfigDir returns the resolved config directory
func (r *Resolver) GetConfigDir() string {
	return r.configDir
}

// GetConfigFile returns the full path to omc.json
func (r *Resolver) GetConfigFile() string {
	return filepath.Join(r.configDir, ConfigFileName)
}

// GetCRDsDir returns the full path to the CRDs directory
func (r *Resolver) GetCRDsDir() string {
	return filepath.Join(r.configDir, CRDsDirName)
}

// GetPullSecretPaths returns possible paths to pull secret files
func (r *Resolver) GetPullSecretPaths() []string {
	return []string{
		filepath.Join(r.configDir, PullSecretTxtName),
		filepath.Join(r.configDir, PullSecretJsonName),
	}
}

// EnsureConfigDir creates the config directory if it doesn't exist
func (r *Resolver) EnsureConfigDir() error {
	if _, err := os.Stat(r.configDir); os.IsNotExist(err) {
		return os.MkdirAll(r.configDir, 0755)
	}
	return nil
}

// EnsureCRDsDir creates the CRDs directory if it doesn't exist
func (r *Resolver) EnsureCRDsDir() error {
	crdsDir := r.GetCRDsDir()
	if _, err := os.Stat(crdsDir); os.IsNotExist(err) {
		return os.MkdirAll(crdsDir, 0755)
	}
	return nil
}

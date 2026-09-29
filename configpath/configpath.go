// Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

package configpath

import (
	"os"
	"path/filepath"
)

const (
	EnvConfigFile        = "OMCCONFIG"
	DefaultConfigDir     = ".omc"
	DefaultConfigFile    = "omc.json"
	CRDsDirName          = "customresourcedefinitions"
	PullSecretTxtName    = "pull-secret.txt"
	PullSecretJsonName   = "pull-secret.json"
)

// Resolver holds the configuration file path and shared resource directory
type Resolver struct {
	configFile string // Path to the config file
	sharedDir  string // Directory for shared resources (CRDs, pull-secrets)
}

// NewResolver creates a resolver with precedence: flag > env > default
// The resolver uses a file-based approach (like KUBECONFIG) where the
// config file can be anywhere, but CRDs and pull-secrets are shared
// in ~/.omc/ across all configs.
func NewResolver(flagValue string) *Resolver {
	var configFile string

	if flagValue != "" {
		// CLI flag takes precedence
		configFile = flagValue
	} else if envValue := os.Getenv(EnvConfigFile); envValue != "" {
		// Environment variable is second priority
		configFile = envValue
	} else {
		// Default to ~/.omc/omc.json
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home cannot be determined
			configFile = filepath.Join(DefaultConfigDir, DefaultConfigFile)
		} else {
			configFile = filepath.Join(home, DefaultConfigDir, DefaultConfigFile)
		}
	}

	// Ensure we have an absolute path for the config file
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		// If we can't get absolute path, use as-is
		absPath = configFile
	}

	// For shared resources, always use ~/.omc/
	home, err := os.UserHomeDir()
	var sharedDir string
	if err != nil {
		sharedDir = DefaultConfigDir
	} else {
		sharedDir = filepath.Join(home, DefaultConfigDir)
	}

	return &Resolver{
		configFile: absPath,
		sharedDir:  sharedDir,
	}
}

// GetConfigFile returns the full path to the config file
func (r *Resolver) GetConfigFile() string {
	return r.configFile
}

// GetConfigDir returns the directory containing the config file
// This is used by viper to set the config path
func (r *Resolver) GetConfigDir() string {
	return filepath.Dir(r.configFile)
}

// GetSharedDir returns the shared directory for global resources
// This is always ~/.omc/ regardless of where the config file is
func (r *Resolver) GetSharedDir() string {
	return r.sharedDir
}

// GetCRDsDir returns the full path to the CRDs directory
// CRDs are shared across all configs, always in ~/.omc/customresourcedefinitions/
func (r *Resolver) GetCRDsDir() string {
	return filepath.Join(r.sharedDir, CRDsDirName)
}

// GetPullSecretPaths returns possible paths to pull secret files
// Pull secrets are shared across all configs, always in ~/.omc/
func (r *Resolver) GetPullSecretPaths() []string {
	return []string{
		filepath.Join(r.sharedDir, PullSecretTxtName),
		filepath.Join(r.sharedDir, PullSecretJsonName),
	}
}

// EnsureConfigDir creates the parent directory of the config file if it doesn't exist
func (r *Resolver) EnsureConfigDir() error {
	configDir := filepath.Dir(r.configFile)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return os.MkdirAll(configDir, 0755)
	}
	return nil
}

// EnsureSharedDir creates the shared directory if it doesn't exist
func (r *Resolver) EnsureSharedDir() error {
	if _, err := os.Stat(r.sharedDir); os.IsNotExist(err) {
		return os.MkdirAll(r.sharedDir, 0755)
	}
	return nil
}

// EnsureCRDsDir creates the CRDs directory if it doesn't exist
// CRDs directory is always in the shared directory
func (r *Resolver) EnsureCRDsDir() error {
	crdsDir := r.GetCRDsDir()
	if _, err := os.Stat(crdsDir); os.IsNotExist(err) {
		return os.MkdirAll(crdsDir, 0755)
	}
	return nil
}

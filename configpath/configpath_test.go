// Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

package configpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolverPrecedence(t *testing.T) {
	// Test flag takes precedence over env
	t.Setenv(EnvConfigHome, "/tmp/env-config")
	resolver := NewResolver("/tmp/flag-config")
	expected := "/tmp/flag-config"
	if resolver.GetConfigDir() != expected {
		t.Errorf("Expected flag to take precedence, got %s, want %s", resolver.GetConfigDir(), expected)
	}

	// Test env takes precedence over default
	resolver = NewResolver("")
	expected = "/tmp/env-config"
	if resolver.GetConfigDir() != expected {
		t.Errorf("Expected env to take precedence, got %s, want %s", resolver.GetConfigDir(), expected)
	}
}

func TestDefaultConfigDir(t *testing.T) {
	// Unset the environment variable
	t.Setenv(EnvConfigHome, "")
	resolver := NewResolver("")

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, DefaultConfigDirName)
	if resolver.GetConfigDir() != expected {
		t.Errorf("Expected %s, got %s", expected, resolver.GetConfigDir())
	}
}

func TestGetConfigFile(t *testing.T) {
	resolver := NewResolver("/tmp/test-config")
	expected := filepath.Join("/tmp/test-config", ConfigFileName)
	if resolver.GetConfigFile() != expected {
		t.Errorf("Expected %s, got %s", expected, resolver.GetConfigFile())
	}
}

func TestGetCRDsDir(t *testing.T) {
	resolver := NewResolver("/tmp/test-config")
	expected := filepath.Join("/tmp/test-config", CRDsDirName)
	if resolver.GetCRDsDir() != expected {
		t.Errorf("Expected %s, got %s", expected, resolver.GetCRDsDir())
	}
}

func TestGetPullSecretPaths(t *testing.T) {
	resolver := NewResolver("/tmp/test-config")
	paths := resolver.GetPullSecretPaths()

	if len(paths) != 2 {
		t.Errorf("Expected 2 pull secret paths, got %d", len(paths))
	}

	expectedTxt := filepath.Join("/tmp/test-config", PullSecretTxtName)
	expectedJson := filepath.Join("/tmp/test-config", PullSecretJsonName)

	if paths[0] != expectedTxt {
		t.Errorf("Expected first path %s, got %s", expectedTxt, paths[0])
	}

	if paths[1] != expectedJson {
		t.Errorf("Expected second path %s, got %s", expectedJson, paths[1])
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "test-omc-config")

	resolver := NewResolver(configDir)

	// Directory should not exist yet
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Errorf("Config directory should not exist yet")
	}

	// Create the directory
	if err := resolver.EnsureConfigDir(); err != nil {
		t.Errorf("Failed to create config directory: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory should exist after EnsureConfigDir()")
	}

	// Calling again should not error
	if err := resolver.EnsureConfigDir(); err != nil {
		t.Errorf("EnsureConfigDir should not error when directory exists: %v", err)
	}
}

func TestEnsureCRDsDir(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "test-omc-config")

	resolver := NewResolver(configDir)

	// Create parent config directory first
	if err := resolver.EnsureConfigDir(); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	crdsDir := resolver.GetCRDsDir()

	// CRDs directory should not exist yet
	if _, err := os.Stat(crdsDir); !os.IsNotExist(err) {
		t.Errorf("CRDs directory should not exist yet")
	}

	// Create the CRDs directory
	if err := resolver.EnsureCRDsDir(); err != nil {
		t.Errorf("Failed to create CRDs directory: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(crdsDir); os.IsNotExist(err) {
		t.Errorf("CRDs directory should exist after EnsureCRDsDir()")
	}

	// Calling again should not error
	if err := resolver.EnsureCRDsDir(); err != nil {
		t.Errorf("EnsureCRDsDir should not error when directory exists: %v", err)
	}
}

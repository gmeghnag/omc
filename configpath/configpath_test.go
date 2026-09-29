// Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

package configpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolverPrecedence(t *testing.T) {
	// Test flag takes precedence over env
	t.Setenv(EnvConfigFile, "/tmp/env-config.json")
	resolver := NewResolver("/tmp/flag-config.json")
	expected := "/tmp/flag-config.json"
	if resolver.GetConfigFile() != expected {
		t.Errorf("Expected flag to take precedence, got %s, want %s", resolver.GetConfigFile(), expected)
	}

	// Test env takes precedence over default
	resolver = NewResolver("")
	expected = "/tmp/env-config.json"
	if resolver.GetConfigFile() != expected {
		t.Errorf("Expected env to take precedence, got %s, want %s", resolver.GetConfigFile(), expected)
	}
}

func TestDefaultConfigFile(t *testing.T) {
	// Unset the environment variable
	t.Setenv(EnvConfigFile, "")
	resolver := NewResolver("")

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, DefaultConfigDir, DefaultConfigFile)
	if resolver.GetConfigFile() != expected {
		t.Errorf("Expected %s, got %s", expected, resolver.GetConfigFile())
	}
}

func TestGetConfigFile(t *testing.T) {
	resolver := NewResolver("/tmp/test-config.json")
	expected := "/tmp/test-config.json"
	if resolver.GetConfigFile() != expected {
		t.Errorf("Expected %s, got %s", expected, resolver.GetConfigFile())
	}
}

func TestGetConfigDir(t *testing.T) {
	resolver := NewResolver("/tmp/configs/case-123.json")
	expected := "/tmp/configs"
	if resolver.GetConfigDir() != expected {
		t.Errorf("Expected %s, got %s", expected, resolver.GetConfigDir())
	}
}

func TestSharedResources(t *testing.T) {
	// Config file can be anywhere
	resolver := NewResolver("/tmp/configs/case-123.json")

	// But CRDs and pull-secrets are always in ~/.omc/
	home, _ := os.UserHomeDir()
	expectedSharedDir := filepath.Join(home, DefaultConfigDir)
	expectedCRDsDir := filepath.Join(expectedSharedDir, CRDsDirName)

	if resolver.GetSharedDir() != expectedSharedDir {
		t.Errorf("Expected shared dir %s, got %s", expectedSharedDir, resolver.GetSharedDir())
	}

	if resolver.GetCRDsDir() != expectedCRDsDir {
		t.Errorf("CRDs should be in shared dir, expected %s, got %s", expectedCRDsDir, resolver.GetCRDsDir())
	}

	// Pull secrets should also be in shared dir
	paths := resolver.GetPullSecretPaths()
	expectedTxt := filepath.Join(expectedSharedDir, PullSecretTxtName)
	expectedJson := filepath.Join(expectedSharedDir, PullSecretJsonName)

	if len(paths) != 2 {
		t.Errorf("Expected 2 pull secret paths, got %d", len(paths))
	}

	if paths[0] != expectedTxt {
		t.Errorf("Expected first path %s, got %s", expectedTxt, paths[0])
	}

	if paths[1] != expectedJson {
		t.Errorf("Expected second path %s, got %s", expectedJson, paths[1])
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "subdir", "test-config.json")

	resolver := NewResolver(configFile)

	// Parent directory should not exist yet
	configDir := filepath.Dir(configFile)
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Errorf("Config parent directory should not exist yet")
	}

	// Create the parent directory
	if err := resolver.EnsureConfigDir(); err != nil {
		t.Errorf("Failed to create config parent directory: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config parent directory should exist after EnsureConfigDir()")
	}

	// Calling again should not error
	if err := resolver.EnsureConfigDir(); err != nil {
		t.Errorf("EnsureConfigDir should not error when directory exists: %v", err)
	}
}

func TestEnsureSharedDir(t *testing.T) {
	// Create resolver with config file in temp location
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.json")

	// Temporarily change HOME to test shared dir creation
	originalHome := os.Getenv("HOME")
	testHome := t.TempDir()
	t.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	resolver := NewResolver(configFile)

	expectedSharedDir := filepath.Join(testHome, DefaultConfigDir)

	// Shared directory should not exist yet
	if _, err := os.Stat(expectedSharedDir); !os.IsNotExist(err) {
		t.Errorf("Shared directory should not exist yet")
	}

	// Create the shared directory
	if err := resolver.EnsureSharedDir(); err != nil {
		t.Errorf("Failed to create shared directory: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(expectedSharedDir); os.IsNotExist(err) {
		t.Errorf("Shared directory should exist after EnsureSharedDir()")
	}

	// Calling again should not error
	if err := resolver.EnsureSharedDir(); err != nil {
		t.Errorf("EnsureSharedDir should not error when directory exists: %v", err)
	}
}

func TestEnsureCRDsDir(t *testing.T) {
	// Create resolver with config file in temp location
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.json")

	// Temporarily change HOME to test CRDs dir creation
	originalHome := os.Getenv("HOME")
	testHome := t.TempDir()
	t.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)

	resolver := NewResolver(configFile)

	// Create shared directory first
	if err := resolver.EnsureSharedDir(); err != nil {
		t.Fatalf("Failed to create shared directory: %v", err)
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

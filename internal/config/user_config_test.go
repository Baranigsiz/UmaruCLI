package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.License != "MIT" {
		t.Errorf("Expected default license 'MIT', got %s", cfg.License)
	}
	if !cfg.GitInit {
		t.Errorf("Expected default GitInit to be true")
	}
}

func TestSetConfigValue_Valid(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("USERPROFILE")
	if origHome == "" {
		origHome = os.Getenv("HOME")
	}
	defer func() {
		_ = os.Setenv("USERPROFILE", origHome)
		_ = os.Setenv("HOME", origHome)
	}()

	_ = os.Setenv("USERPROFILE", tempDir)
	_ = os.Setenv("HOME", tempDir)

	// Set package manager
	cfg, err := SetConfigValue("package-manager", "pnpm")
	if err != nil {
		t.Fatalf("SetConfigValue failed: %v", err)
	}
	if cfg.PackageManager != "pnpm" {
		t.Errorf("Expected PackageManager 'pnpm', got '%s'", cfg.PackageManager)
	}

	// Verify it persisted
	loaded := LoadUserConfig()
	if loaded.PackageManager != "pnpm" {
		t.Errorf("Expected loaded PackageManager 'pnpm', got '%s'", loaded.PackageManager)
	}

	// Set author
	cfg, err = SetConfigValue("author", "Baran Igsiz")
	if err != nil {
		t.Fatalf("SetConfigValue author failed: %v", err)
	}
	if cfg.Author != "Baran Igsiz" {
		t.Errorf("Expected Author 'Baran Igsiz', got '%s'", cfg.Author)
	}

	// Reset
	err = ResetConfig()
	if err != nil {
		t.Fatalf("ResetConfig failed: %v", err)
	}
	resetCfg := LoadUserConfig()
	if resetCfg.PackageManager != "" || resetCfg.Author != "" {
		t.Errorf("Expected empty values after reset, got %+v", resetCfg)
	}
}

func TestSetConfigValue_Invalid(t *testing.T) {
	_, err := SetConfigValue("invalid-key-xyz", "value")
	if err == nil {
		t.Errorf("Expected error for invalid config key, got nil")
	}

	_, err = SetConfigValue("package-manager", "invalid-pm")
	if err == nil {
		t.Errorf("Expected error for invalid package manager, got nil")
	}
}

func TestGetConfigFilePath(t *testing.T) {
	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath failed: %v", err)
	}
	if filepath.Base(path) != ".umarurc.json" {
		t.Errorf("Expected filename .umarurc.json, got %s", path)
	}
}

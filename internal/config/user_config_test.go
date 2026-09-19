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

func TestSetCustomConfigFile(t *testing.T) {
	customPath := filepath.Join(t.TempDir(), "custom-umaru.json")
	SetCustomConfigFile(customPath)
	defer SetCustomConfigFile("")

	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath failed: %v", err)
	}
	if path != customPath {
		t.Errorf("Expected path %s, got %s", customPath, path)
	}
}

func TestSetTestConfigDir(t *testing.T) {
	tempDir := t.TempDir()
	SetTestConfigDir(tempDir)
	defer SetTestConfigDir("")

	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath failed: %v", err)
	}
	expected := filepath.Join(tempDir, ".umarurc.json")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestGetConfigFilePath_Env(t *testing.T) {
	tempDir := t.TempDir()
	origEnv := os.Getenv("UMARU_CONFIG_DIR")
	_ = os.Setenv("UMARU_CONFIG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("UMARU_CONFIG_DIR", origEnv)
	}()

	// Ensure custom config file and test dir are empty
	SetCustomConfigFile("")
	SetTestConfigDir("")

	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath failed with env: %v", err)
	}
	expected := filepath.Join(tempDir, ".umarurc.json")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestUnsetConfigValue(t *testing.T) {
	tempDir := t.TempDir()
	SetTestConfigDir(tempDir)
	defer SetTestConfigDir("")

	// Set initial values
	_, _ = SetConfigValue("package-manager", "yarn")
	_, _ = SetConfigValue("author", "Jane Doe")
	_, _ = SetConfigValue("license", "GPL-3.0")
	_, _ = SetConfigValue("git-init", "false")

	// Unset package-manager
	cfg, err := UnsetConfigValue("package-manager")
	if err != nil {
		t.Fatalf("UnsetConfigValue package-manager failed: %v", err)
	}
	if cfg.PackageManager != DefaultConfig().PackageManager {
		t.Errorf("Expected default PackageManager, got %s", cfg.PackageManager)
	}

	// Unset author
	cfg, err = UnsetConfigValue("author")
	if err != nil {
		t.Fatalf("UnsetConfigValue author failed: %v", err)
	}
	if cfg.Author != DefaultConfig().Author {
		t.Errorf("Expected default Author, got %s", cfg.Author)
	}

	// Unset license
	cfg, err = UnsetConfigValue("license")
	if err != nil {
		t.Fatalf("UnsetConfigValue license failed: %v", err)
	}
	if cfg.License != DefaultConfig().License {
		t.Errorf("Expected default License, got %s", cfg.License)
	}

	// Unset git-init
	cfg, err = UnsetConfigValue("git-init")
	if err != nil {
		t.Fatalf("UnsetConfigValue git-init failed: %v", err)
	}
	if cfg.GitInit != DefaultConfig().GitInit {
		t.Errorf("Expected default GitInit, got %v", cfg.GitInit)
	}

	// Test aliases
	_, _ = SetConfigValue("package-manager", "bun")
	cfg, err = UnsetConfigValue("pm")
	if err != nil || cfg.PackageManager != "" {
		t.Errorf("UnsetConfigValue with alias 'pm' failed: %v", err)
	}

	_, _ = SetConfigValue("git-init", "false")
	cfg, err = UnsetConfigValue("git")
	if err != nil || !cfg.GitInit {
		t.Errorf("UnsetConfigValue with alias 'git' failed: %v", err)
	}

	// Test invalid key
	_, err = UnsetConfigValue("non-existent-key")
	if err == nil {
		t.Errorf("Expected error for non-existent key, got nil")
	}
}

func TestSetConfigValue_Booleans(t *testing.T) {
	tempDir := t.TempDir()
	SetTestConfigDir(tempDir)
	defer SetTestConfigDir("")

	validTrue := []string{"true", "1", "yes"}
	for _, val := range validTrue {
		cfg, err := SetConfigValue("git-init", val)
		if err != nil {
			t.Errorf("Expected valid boolean for %s, got error: %v", val, err)
		}
		if !cfg.GitInit {
			t.Errorf("Expected GitInit to be true for %s", val)
		}
	}

	validFalse := []string{"false", "0", "no"}
	for _, val := range validFalse {
		cfg, err := SetConfigValue("git-init", val)
		if err != nil {
			t.Errorf("Expected valid boolean for %s, got error: %v", val, err)
		}
		if cfg.GitInit {
			t.Errorf("Expected GitInit to be false for %s", val)
		}
	}

	// Invalid boolean
	_, err := SetConfigValue("git-init", "not-a-bool")
	if err == nil {
		t.Errorf("Expected error for invalid boolean 'not-a-bool', got nil")
	}
}

func TestLoadUserConfig_CorruptFile(t *testing.T) {
	tempDir := t.TempDir()
	SetTestConfigDir(tempDir)
	defer SetTestConfigDir("")

	cfgPath := filepath.Join(tempDir, ".umarurc.json")
	_ = os.WriteFile(cfgPath, []byte("{invalid-json}"), 0644)

	loaded := LoadUserConfig()
	def := DefaultConfig()
	if loaded.License != def.License || loaded.GitInit != def.GitInit {
		t.Errorf("Expected fallback to DefaultConfig on corrupt JSON, got %+v", loaded)
	}
}


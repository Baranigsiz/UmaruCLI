package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UserConfig represents the persistent user preferences saved in ~/.umarurc.json
type UserConfig struct {
	PackageManager string `json:"packageManager,omitempty"` // "npm", "pnpm", "yarn", "bun"
	Author         string `json:"author,omitempty"`         // Default author name
	License        string `json:"license,omitempty"`        // Default license (e.g. "MIT")
	GitInit        bool   `json:"gitInit"`                  // Auto git init (default: true)
	Telemetry      bool   `json:"telemetry"`                // Reserved for future opt-in metrics
}

// DefaultConfig returns the default configuration
func DefaultConfig() UserConfig {
	return UserConfig{
		PackageManager: "",
		Author:         "",
		License:        "MIT",
		GitInit:        true,
		Telemetry:      false,
	}
}

// GetConfigFilePath returns the absolute path to ~/.umarurc.json
func GetConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}
	return filepath.Join(home, ".umarurc.json"), nil
}

// LoadUserConfig reads and parses the configuration file. If it doesn't exist, returns default config.
func LoadUserConfig() UserConfig {
	cfg := DefaultConfig()

	path, err := GetConfigFilePath()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}

// SaveUserConfig writes the given configuration to ~/.umarurc.json
func SaveUserConfig(cfg UserConfig) error {
	path, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode user config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save config to %s: %w", path, err)
	}

	return nil
}

// SetConfigValue updates a specific key in the user configuration
func SetConfigValue(key string, value string) (UserConfig, error) {
	cfg := LoadUserConfig()
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)

	switch key {
	case "package-manager", "pm", "packagemanager":
		val := strings.ToLower(value)
		if val != "npm" && val != "pnpm" && val != "yarn" && val != "bun" && val != "" {
			return cfg, fmt.Errorf("invalid package manager '%s'. Supported: npm, pnpm, yarn, bun", value)
		}
		cfg.PackageManager = val
	case "author":
		cfg.Author = value
	case "license":
		cfg.License = value
	case "git-init", "gitinit", "git":
		val := strings.ToLower(value)
		cfg.GitInit = (val == "true" || val == "1" || val == "yes")
	default:
		return cfg, fmt.Errorf("unknown config key '%s'. Supported keys: package-manager, author, license, git-init", key)
	}

	if err := SaveUserConfig(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// ResetConfig resets ~/.umarurc.json to default values
func ResetConfig() error {
	return SaveUserConfig(DefaultConfig())
}

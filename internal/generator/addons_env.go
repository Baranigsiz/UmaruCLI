package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvBlock represents a group of environment variables related to an addon
type EnvBlock struct {
	Comment string
	Vars    [][2]string // [key, value]
}

func getAddonEnvBlocks(config ProjectConfig) []EnvBlock {
	var blocks []EnvBlock

	db := strings.ToLower(strings.TrimSpace(config.Addons.Database))
	if db == "postgres" {
		dbName := config.SafeName
		if dbName == "" {
			dbName = "app_db"
		}
		blocks = append(blocks, EnvBlock{
			Comment: "Database (PostgreSQL)",
			Vars: [][2]string{
				{"DB_HOST", "localhost"},
				{"DB_PORT", "5432"},
				{"DB_USER", "postgres"},
				{"DB_PASSWORD", "postgres"},
				{"DB_NAME", dbName},
				{"DB_SSLMODE", "disable"},
			},
		})
	} else if db == "sqlite" {
		blocks = append(blocks, EnvBlock{
			Comment: "Database (SQLite)",
			Vars: [][2]string{
				{"DB_PATH", "./app.db"},
			},
		})
	}

	auth := strings.ToLower(strings.TrimSpace(config.Addons.Auth))
	if auth == "jwt" {
		blocks = append(blocks, EnvBlock{
			Comment: "Authentication (JWT)",
			Vars: [][2]string{
				{"JWT_SECRET", "super-secret-key-change-in-production"},
				{"JWT_EXPIRES_IN", "24h"},
			},
		})
	}

	if config.Addons.Redis {
		blocks = append(blocks, EnvBlock{
			Comment: "Cache (Redis)",
			Vars: [][2]string{
				{"REDIS_ADDR", "localhost:6379"},
				{"REDIS_HOST", "localhost"},
				{"REDIS_PORT", "6379"},
				{"REDIS_PASSWORD", ""},
				{"REDIS_DB", "0"},
			},
		})
	}

	return blocks
}

// injectEnvVariables ensures all selected addon variables are documented in .env.example
// and copies to .env if .env does not already exist.
func injectEnvVariables(baseDir string, config ProjectConfig) error {
	blocks := getAddonEnvBlocks(config)
	if len(blocks) == 0 {
		return nil
	}

	envExamplePath := filepath.Join(baseDir, ".env.example")
	envPath := filepath.Join(baseDir, ".env")

	var currentContent string
	if data, err := os.ReadFile(envExamplePath); err == nil {
		data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
		currentContent = string(data)
	}

	existingKeys := make(map[string]bool)
	for _, line := range strings.Split(currentContent, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) >= 1 {
				key := strings.TrimSpace(parts[0])
				existingKeys[key] = true
			}
		}
	}

	var toAppend strings.Builder
	for _, block := range blocks {
		var missingVars [][2]string
		for _, pair := range block.Vars {
			if !existingKeys[pair[0]] {
				missingVars = append(missingVars, pair)
			}
		}

		if len(missingVars) > 0 {
			if toAppend.Len() > 0 || len(currentContent) > 0 {
				toAppend.WriteString("\n")
			}
			toAppend.WriteString(fmt.Sprintf("# %s\n", block.Comment))
			for _, pair := range missingVars {
				toAppend.WriteString(fmt.Sprintf("%s=%s\n", pair[0], pair[1]))
			}
		}
	}

	if toAppend.Len() > 0 {
		newContent := currentContent
		if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
			newContent += "\n"
		}
		newContent += toAppend.String()

		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(envExamplePath, []byte(newContent), 0644); err != nil {
			return err
		}
	}

	// If .env does not exist, copy .env.example to .env
	if !fileExists(envPath) && fileExists(envExamplePath) {
		if content, err := os.ReadFile(envExamplePath); err == nil {
			_ = os.WriteFile(envPath, content, 0644)
		}
	}

	return nil
}

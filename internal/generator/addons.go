package generator

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// AddonConfig holds optional infrastructure and middleware add-ons
type AddonConfig struct {
	Database string `json:"database,omitempty"` // "none", "postgres", "sqlite"
	Auth     string `json:"auth,omitempty"`     // "none", "jwt"
	Redis    bool   `json:"redis,omitempty"`    // true/false
	Docker   bool   `json:"docker,omitempty"`   // true/false
}

// TemplateSupportsAddons checks if the template supports the optional addons (DB, Auth, Redis, Docker)
func TemplateSupportsAddons(templateID string) bool {
	return strings.HasPrefix(templateID, "go-") ||
		templateID == "fullstack-go-react" ||
		templateID == "fullstack-ts-monorepo" ||
		strings.HasPrefix(templateID, "node-") ||
		strings.HasPrefix(templateID, "nestjs-") ||
		strings.HasPrefix(templateID, "hono-") ||
		strings.HasPrefix(templateID, "fastify-") ||
		strings.HasPrefix(templateID, "bun-") ||
		strings.HasPrefix(templateID, "python-") ||
		strings.HasPrefix(templateID, "ai-")
}

// HasAddons returns true if any addon is enabled
func (a AddonConfig) HasAddons() bool {
	db := strings.ToLower(strings.TrimSpace(a.Database))
	auth := strings.ToLower(strings.TrimSpace(a.Auth))
	return (db != "" && db != "none") || (auth != "" && auth != "none") || a.Redis || a.Docker
}

// GetAddonFiles returns the list of file paths that will be generated for the selected addons
func GetAddonFiles(config ProjectConfig) []string {
	var files []string
	if !config.Addons.HasAddons() {
		return files
	}

	baseDir := getAddonBaseDir(config)

	if dbFiles := getDatabaseFiles(config, baseDir); len(dbFiles) > 0 {
		files = append(files, dbFiles...)
	}
	if authFiles := getAuthFiles(config, baseDir); len(authFiles) > 0 {
		files = append(files, authFiles...)
	}
	if redisFiles := getRedisFiles(config, baseDir); len(redisFiles) > 0 {
		files = append(files, redisFiles...)
	}
	if config.Addons.Docker {
		files = append(files, getDockerFiles(baseDir)...)
	}

	return files
}

// GenerateAddons writes the addon template files into the target project
func GenerateAddons(config ProjectConfig) error {
	if !config.Addons.HasAddons() {
		return nil
	}

	baseDir := getAddonBaseDir(config)

	if err := generateDatabaseAddon(config, baseDir); err != nil {
		return err
	}
	if err := generateAuthAddon(config, baseDir); err != nil {
		return err
	}
	if err := generateRedisAddon(config, baseDir); err != nil {
		return err
	}
	if config.Addons.Docker {
		if err := generateDockerAddon(config, baseDir); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions shared across addon generators

func getAddonBaseDir(config ProjectConfig) string {
	if config.Template == "fullstack-go-react" || config.Template == "fullstack-ts-monorepo" {
		return filepath.Join(config.TargetDir, "apps", "api")
	}
	return config.TargetDir
}

func writeAddonFile(baseDir, relPath, content string) error {
	fullPath := filepath.Join(baseDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

func isGoTemplate(templateID string) bool {
	return strings.HasPrefix(templateID, "go-") || templateID == "fullstack-go-react"
}

func isNodeTemplate(templateID string) bool {
	return strings.HasPrefix(templateID, "node-") ||
		strings.HasPrefix(templateID, "nestjs-") ||
		strings.HasPrefix(templateID, "hono-") ||
		strings.HasPrefix(templateID, "fastify-") ||
		strings.HasPrefix(templateID, "bun-") ||
		templateID == "fullstack-ts-monorepo"
}

func isPythonTemplate(templateID string) bool {
	return strings.HasPrefix(templateID, "python-") || strings.HasPrefix(templateID, "ai-")
}

// injectPythonDependencies appends missing packages to requirements.txt
func injectPythonDependencies(requirementsPath string, packages []string) error {
	content, err := os.ReadFile(requirementsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))

	lines := strings.Split(string(content), "\n")
	existing := make(map[string]bool)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			pkgName := strings.Split(strings.Split(trimmed, ">")[0], "=")[0]
			pkgName = strings.Split(pkgName, "[")[0]
			existing[strings.ToLower(strings.TrimSpace(pkgName))] = true
		}
	}

	var toAdd []string
	for _, pkg := range packages {
		pkgName := strings.Split(strings.Split(pkg, ">")[0], "=")[0]
		pkgName = strings.Split(pkgName, "[")[0]
		if !existing[strings.ToLower(strings.TrimSpace(pkgName))] {
			toAdd = append(toAdd, pkg)
		}
	}

	if len(toAdd) > 0 {
		newContent := string(content)
		if !strings.HasSuffix(newContent, "\n") && len(newContent) > 0 {
			newContent += "\n"
		}
		newContent += strings.Join(toAdd, "\n") + "\n"
		return os.WriteFile(requirementsPath, []byte(newContent), 0644)
	}

	return nil
}

// injectNodeDependencies safely injects dependencies and devDependencies into package.json
func injectNodeDependencies(packageJSONPath string, deps map[string]string, devDeps map[string]string) error {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	var pkgMap map[string]interface{}
	if err := json.Unmarshal(data, &pkgMap); err != nil {
		return err
	}

	if len(deps) > 0 {
		var currentDeps map[string]interface{}
		if existing, ok := pkgMap["dependencies"].(map[string]interface{}); ok {
			currentDeps = existing
		} else {
			currentDeps = make(map[string]interface{})
		}
		for k, v := range deps {
			if _, exists := currentDeps[k]; !exists {
				currentDeps[k] = v
			}
		}
		pkgMap["dependencies"] = currentDeps
	}

	if len(devDeps) > 0 {
		var currentDevDeps map[string]interface{}
		if existing, ok := pkgMap["devDependencies"].(map[string]interface{}); ok {
			currentDevDeps = existing
		} else {
			currentDevDeps = make(map[string]interface{})
		}
		for k, v := range devDeps {
			if _, exists := currentDevDeps[k]; !exists {
				currentDevDeps[k] = v
			}
		}
		pkgMap["devDependencies"] = currentDevDeps
	}

	updated, err := json.MarshalIndent(pkgMap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(packageJSONPath, append(updated, '\n'), 0644)
}

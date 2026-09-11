package generator

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"umaru/internal/templates"
)

// TestSmoke_AllTemplatesRenderCleanly ensures all 20 templates generate cleanly
// without unrendered template tags, leaked files, or invalid paths.
func TestSmoke_AllTemplatesRenderCleanly(t *testing.T) {
	allTemplates, err := templates.GetAvailableTemplates()
	if err != nil {
		t.Fatalf("Failed to fetch available templates: %v", err)
	}

	if len(allTemplates) < 23 {
		t.Errorf("Expected at least 23 templates, found %d", len(allTemplates))
	}

	for _, tmpl := range allTemplates {
		t.Run("Smoke_"+tmpl.ID, func(t *testing.T) {
			tempDir := t.TempDir()
			targetPath := filepath.Join(tempDir, "test-app")

			cfg, err := ResolveProjectConfig(targetPath, tmpl.ID)
			if err != nil {
				t.Fatalf("ResolveProjectConfig failed: %v", err)
			}
			cfg.ProjectName = "Smoke Test App"
			cfg.SafeName = "smoke-test-app"
			cfg.ModuleName = "smoke-test-app"
			cfg.Author = "Test Author"
			cfg.License = "MIT"

			if err := Generate(cfg); err != nil {
				t.Fatalf("Generate failed for template %s: %v", tmpl.ID, err)
			}

			// Verify all generated files
			err = filepath.WalkDir(targetPath, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				rel, _ := filepath.Rel(targetPath, p)

				// 1. No .tmpl files should remain
				if strings.HasSuffix(p, ".tmpl") {
					t.Errorf("[%s] File %s still has .tmpl extension", tmpl.ID, rel)
				}

				// 2. template.json should never be leaked
				if rel == "template.json" {
					t.Errorf("[%s] template.json was copied to generated project", tmpl.ID)
				}

				// 3. Check for unrendered template tags in text files
				ext := filepath.Ext(p)
				textExtensions := map[string]bool{
					".go": true, ".json": true, ".toml": true, ".md": true,
					".yml": true, ".yaml": true, ".ts": true, ".tsx": true,
					".js": true, ".jsx": true, ".py": true, ".html": true,
					".env": true, ".example": true,
				}

				// For templates that ship runtime HTML templates (e.g. go-htmx views),
				// skip checking runtime template tags inside views/
				if tmpl.ID == "go-htmx" && strings.HasPrefix(filepath.ToSlash(rel), "views/") {
					return nil
				}

				if textExtensions[ext] || strings.HasPrefix(filepath.Base(p), ".env") {
					content, readErr := os.ReadFile(p)
					if readErr == nil {
						str := string(content)
						if strings.Contains(str, "{{.") || strings.Contains(str, "}}") {
							t.Errorf("[%s] Unrendered template tag found in %s:\n%s", tmpl.ID, rel, str)
						}
					}
				}

				return nil
			})

			if err != nil {
				t.Fatalf("[%s] Failed to walk generated project: %v", tmpl.ID, err)
			}
		})
	}
}

// TestSmoke_ManifestValidity verifies that generated manifests (package.json, go.mod, Cargo.toml) are syntactically valid.
func TestSmoke_ManifestValidity(t *testing.T) {
	allTemplates, err := templates.GetAvailableTemplates()
	if err != nil {
		t.Fatalf("Failed to fetch available templates: %v", err)
	}

	for _, tmpl := range allTemplates {
		t.Run("Manifest_"+tmpl.ID, func(t *testing.T) {
			tempDir := t.TempDir()
			targetPath := filepath.Join(tempDir, "app")

			cfg, err := ResolveProjectConfig(targetPath, tmpl.ID)
			if err != nil {
				t.Fatalf("ResolveProjectConfig failed: %v", err)
			}
			cfg.SafeName = "manifest-test-app"
			cfg.ModuleName = "manifest-test-app"

			if err := Generate(cfg); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			// Check package.json if present
			pkgJSONPath := filepath.Join(targetPath, "package.json")
			if _, err := os.Stat(pkgJSONPath); err == nil {
				data, err := os.ReadFile(pkgJSONPath)
				if err != nil {
					t.Fatalf("[%s] Failed to read package.json: %v", tmpl.ID, err)
				}
				var parsed map[string]interface{}
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Errorf("[%s] package.json is invalid JSON: %v", tmpl.ID, err)
				}
				if name, ok := parsed["name"].(string); !ok || name == "" {
					t.Errorf("[%s] package.json is missing 'name' field", tmpl.ID)
				}
			}

			// Check go.mod if present
			goModPath := filepath.Join(targetPath, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				data, err := os.ReadFile(goModPath)
				if err != nil {
					t.Fatalf("[%s] Failed to read go.mod: %v", tmpl.ID, err)
				}
				content := string(data)
				if !strings.Contains(content, "module ") {
					t.Errorf("[%s] go.mod missing 'module' declaration", tmpl.ID)
				}
				if !strings.Contains(content, "go 1.") {
					t.Errorf("[%s] go.mod missing 'go 1.x' declaration", tmpl.ID)
				}
			}

			// Check Cargo.toml if present
			cargoPath := filepath.Join(targetPath, "Cargo.toml")
			if _, err := os.Stat(cargoPath); err == nil {
				data, err := os.ReadFile(cargoPath)
				if err != nil {
					t.Fatalf("[%s] Failed to read Cargo.toml: %v", tmpl.ID, err)
				}
				content := string(data)
				if !strings.Contains(content, "[package]") {
					t.Errorf("[%s] Cargo.toml missing [package] table", tmpl.ID)
				}
				if !strings.Contains(content, "name =") {
					t.Errorf("[%s] Cargo.toml missing 'name' key", tmpl.ID)
				}
			}
		})
	}
}

// TestSmoke_FullAddonStackIntegration verifies that a project generated with all addons
// (Database, Auth, Redis, Docker, CI) produces a complete, cohesive project.
func TestSmoke_FullAddonStackIntegration(t *testing.T) {
	testTemplates := []string{"go-fiber", "node-express", "python-fastapi", "bun-elysia"}

	for _, tmplID := range testTemplates {
		t.Run("FullStack_"+tmplID, func(t *testing.T) {
			tempDir := t.TempDir()
			targetPath := filepath.Join(tempDir, "full-app")

			cfg, err := ResolveProjectConfig(targetPath, tmplID)
			if err != nil {
				t.Fatalf("ResolveProjectConfig failed: %v", err)
			}
			cfg.SafeName = "full-stack-app"
			cfg.ModuleName = "full-stack-app"
			cfg.Addons = AddonConfig{
				Database: "postgres",
				Auth:     "jwt",
				Redis:    true,
				Docker:   true,
				CI:       true,
			}

			if err := Generate(cfg); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			if err := GenerateAddons(cfg); err != nil {
				t.Fatalf("GenerateAddons failed: %v", err)
			}

			// 1. Dockerfile must exist
			dockerfilePath := filepath.Join(targetPath, "Dockerfile")
			if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
				t.Errorf("[%s] Expected Dockerfile to be generated", tmplID)
			}

			// 2. docker-compose.yml must exist and include postgres & redis
			composePath := filepath.Join(targetPath, "docker-compose.yml")
			composeData, err := os.ReadFile(composePath)
			if err != nil {
				t.Fatalf("[%s] Failed to read docker-compose.yml: %v", tmplID, err)
			}
			composeStr := string(composeData)
			if !strings.Contains(composeStr, "postgres:16-alpine") {
				t.Errorf("[%s] docker-compose.yml missing postgres service", tmplID)
			}
			if !strings.Contains(composeStr, "redis:7-alpine") {
				t.Errorf("[%s] docker-compose.yml missing redis service", tmplID)
			}

			// 3. .env and .env.example must exist and contain environment variables
			envExamplePath := filepath.Join(targetPath, ".env.example")
			envData, err := os.ReadFile(envExamplePath)
			if err != nil {
				t.Fatalf("[%s] Failed to read .env.example: %v", tmplID, err)
			}
			envStr := string(envData)
			if !strings.Contains(envStr, "DB_HOST") {
				t.Errorf("[%s] .env.example missing DB_HOST", tmplID)
			}
			if !strings.Contains(envStr, "JWT_SECRET") {
				t.Errorf("[%s] .env.example missing JWT_SECRET", tmplID)
			}
			if !strings.Contains(envStr, "REDIS_ADDR") && !strings.Contains(envStr, "REDIS_HOST") {
				t.Errorf("[%s] .env.example missing REDIS config", tmplID)
			}

			// 4. CI workflow must exist
			ciPath := filepath.Join(targetPath, ".github", "workflows", "ci.yml")
			if _, err := os.Stat(ciPath); os.IsNotExist(err) {
				t.Errorf("[%s] Expected .github/workflows/ci.yml to be generated", tmplID)
			}
		})
	}
}

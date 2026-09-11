package generator

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectType string

const (
	ProjectTypeGo      ProjectType = "go"
	ProjectTypeNode    ProjectType = "node"
	ProjectTypePython  ProjectType = "python"
	ProjectTypeRust    ProjectType = "rust"
	ProjectTypeUnknown ProjectType = "unknown"
)

type DetectedProject struct {
	Type        ProjectType // "go", "node", "python", "rust"
	Framework   string      // "go-fiber", "go-gin", "go-echo", "node-express", "fastify-api", "hono-api", "nestjs-api", "python-fastapi", "rust-axum", "rust-actix"
	TargetDir   string      // Cleaned target directory
	ModuleName  string      // Go module name or safe identifier
	ProjectName string      // Base directory or package name
}

// ToProjectConfig converts a detected project into a ProjectConfig for addon generation
func (p *DetectedProject) ToProjectConfig(addons AddonConfig) ProjectConfig {
	return ProjectConfig{
		ProjectName: p.ProjectName,
		SafeName:    Slugify(p.ProjectName),
		ModuleName:  p.ModuleName,
		TargetDir:   p.TargetDir,
		Template:    p.Framework,
		Addons:      addons,
	}
}

// DetectProject inspects the given directory to determine the language and framework
func DetectProject(dir string) (*DetectedProject, error) {
	targetDir := filepath.Clean(dir)
	info, err := os.Stat(targetDir)
	if err != nil {
		return nil, fmt.Errorf("directory '%s' does not exist: %w", targetDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path '%s' is not a directory", targetDir)
	}

	baseName := filepath.Base(targetDir)
	if baseName == "." || baseName == "/" || baseName == "\\" {
		absDir, err := filepath.Abs(targetDir)
		if err == nil {
			baseName = filepath.Base(absDir)
		}
	}

	// 0. Check for Fullstack Monorepos (apps/api and apps/web)
	appsApiDir := filepath.Join(targetDir, "apps", "api")
	appsWebDir := filepath.Join(targetDir, "apps", "web")
	if isDir(appsApiDir) && isDir(appsWebDir) {
		if fileExists(filepath.Join(appsApiDir, "go.mod")) {
			modData, _ := os.ReadFile(filepath.Join(appsApiDir, "go.mod"))
			moduleName := extractGoModuleName(string(modData))
			if moduleName == "" {
				moduleName = Slugify(baseName)
			}
			return &DetectedProject{
				Type:        ProjectTypeGo,
				Framework:   "fullstack-go-react",
				TargetDir:   targetDir,
				ModuleName:  moduleName,
				ProjectName: baseName,
			}, nil
		}

		if fileExists(filepath.Join(appsApiDir, "package.json")) {
			return &DetectedProject{
				Type:        ProjectTypeNode,
				Framework:   "fullstack-ts-monorepo",
				TargetDir:   targetDir,
				ModuleName:  Slugify(baseName),
				ProjectName: baseName,
			}, nil
		}
	}

	// 1. Check for Go (go.mod)
	goModPath := filepath.Join(targetDir, "go.mod")
	if modData, err := os.ReadFile(goModPath); err == nil {
		modData = bytes.TrimPrefix(modData, []byte("\xef\xbb\xbf"))
		modContent := string(modData)
		moduleName := extractGoModuleName(modContent)
		if moduleName == "" {
			moduleName = Slugify(baseName)
		}

		framework := "go-fiber"
		if strings.Contains(modContent, "github.com/spf13/cobra") {
			framework = "go-cli"
		} else if strings.Contains(modContent, "github.com/charmbracelet/bubbletea") {
			framework = "go-tui"
		} else if strings.Contains(modContent, "github.com/gofiber/template/html") {
			framework = "go-htmx"
		} else if strings.Contains(modContent, "github.com/labstack/echo") {
			framework = "go-echo"
		} else if strings.Contains(modContent, "github.com/gin-gonic/gin") {
			framework = "go-gin"
		} else if strings.Contains(modContent, "github.com/gofiber/fiber") {
			framework = "go-fiber"
		}

		return &DetectedProject{
			Type:        ProjectTypeGo,
			Framework:   framework,
			TargetDir:   targetDir,
			ModuleName:  moduleName,
			ProjectName: baseName,
		}, nil
	}

	// 2. Check for Node.js (package.json)
	pkgJSONPath := filepath.Join(targetDir, "package.json")
	if pkgData, err := os.ReadFile(pkgJSONPath); err == nil {
		pkgData = bytes.TrimPrefix(pkgData, []byte("\xef\xbb\xbf"))
		var pkgMap map[string]interface{}
		_ = json.Unmarshal(pkgData, &pkgMap)

		projectName := baseName
		if name, ok := pkgMap["name"].(string); ok && strings.TrimSpace(name) != "" {
			projectName = strings.TrimSpace(name)
		}

		allDeps := make(map[string]bool)
		if deps, ok := pkgMap["dependencies"].(map[string]interface{}); ok {
			for k := range deps {
				allDeps[strings.ToLower(k)] = true
			}
		}
		if devDeps, ok := pkgMap["devDependencies"].(map[string]interface{}); ok {
			for k := range devDeps {
				allDeps[strings.ToLower(k)] = true
			}
		}

		framework := "node-express"
		if allDeps["@tauri-apps/api"] || allDeps["@tauri-apps/cli"] {
			framework = "tauri-desktop"
		} else if allDeps["elysia"] {
			framework = "bun-elysia"
		} else if allDeps["wrangler"] || allDeps["@cloudflare/workers-types"] {
			framework = "hono-cloudflare"
		} else if allDeps["hono"] {
			framework = "hono-api"
		} else if allDeps["fastify"] {
			framework = "fastify-api"
		} else if allDeps["@nestjs/core"] {
			framework = "nestjs-api"
		} else if allDeps["express"] {
			framework = "node-express"
		} else if allDeps["next"] {
			framework = "nextjs-tailwind"
		} else if allDeps["astro"] {
			framework = "astro-tailwind"
		} else if allDeps["svelte"] {
			framework = "svelte-vite-ts"
		} else if allDeps["vue"] {
			framework = "vue-vite-ts"
		} else if allDeps["react"] {
			framework = "react-vite-ts"
		}

		return &DetectedProject{
			Type:        ProjectTypeNode,
			Framework:   framework,
			TargetDir:   targetDir,
			ModuleName:  Slugify(projectName),
			ProjectName: projectName,
		}, nil
	}

	// 3. Check for Python (requirements.txt or pyproject.toml)
	reqPath := filepath.Join(targetDir, "requirements.txt")
	pyprojPath := filepath.Join(targetDir, "pyproject.toml")
	hasReq := fileExists(reqPath)
	hasPyproj := fileExists(pyprojPath)

	if hasReq || hasPyproj {
		framework := "python-fastapi"
		if hasReq {
			if data, err := os.ReadFile(reqPath); err == nil {
				reqStr := string(data)
				if strings.Contains(reqStr, "chromadb") || strings.Contains(reqStr, "google-generativeai") {
					framework = "ai-rag-agent"
				}
			}
		}

		return &DetectedProject{
			Type:        ProjectTypePython,
			Framework:   framework,
			TargetDir:   targetDir,
			ModuleName:  Slugify(baseName),
			ProjectName: baseName,
		}, nil
	}

	// 4. Check for Rust (Cargo.toml)
	cargoPath := filepath.Join(targetDir, "Cargo.toml")
	if cargoData, err := os.ReadFile(cargoPath); err == nil {
		cargoData = bytes.TrimPrefix(cargoData, []byte("\xef\xbb\xbf"))
		cargoContent := string(cargoData)
		pkgName := extractCargoPackageName(cargoContent)
		if pkgName == "" {
			pkgName = baseName
		}

		framework := "rust-axum"
		if strings.Contains(cargoContent, "actix-web") {
			framework = "rust-actix"
		} else if strings.Contains(cargoContent, "axum") {
			framework = "rust-axum"
		}

		return &DetectedProject{
			Type:        ProjectTypeRust,
			Framework:   framework,
			TargetDir:   targetDir,
			ModuleName:  Slugify(pkgName),
			ProjectName: pkgName,
		}, nil
	}

	return nil, fmt.Errorf("no supported project found in '%s' (must contain go.mod, package.json, requirements.txt/pyproject.toml, or Cargo.toml)", targetDir)
}

func extractCargoPackageName(cargoContent string) string {
	scanner := bufio.NewScanner(strings.NewReader(cargoContent))
	inPackage := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inPackage = (line == "[package]")
			continue
		}
		if inPackage && strings.HasPrefix(line, "name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				if cIdx := strings.Index(val, "#"); cIdx != -1 {
					val = strings.TrimSpace(val[:cIdx])
				}
				val = strings.Trim(val, `"'`)
				if val != "" {
					return val
				}
			}
		}
	}
	return ""
}

func extractGoModuleName(modContent string) string {
	scanner := bufio.NewScanner(strings.NewReader(modContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}



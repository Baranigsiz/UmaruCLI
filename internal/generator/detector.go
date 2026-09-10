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
	ProjectTypeUnknown ProjectType = "unknown"
)

type DetectedProject struct {
	Type        ProjectType // "go", "node", "python"
	Framework   string      // "go-fiber", "go-gin", "go-echo", "node-express", "fastify-api", "hono-api", "nestjs-api", "python-fastapi"
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
		if allDeps["elysia"] {
			framework = "bun-elysia"
		} else if allDeps["hono"] {
			framework = "hono-api"
		} else if allDeps["fastify"] {
			framework = "fastify-api"
		} else if allDeps["@nestjs/core"] {
			framework = "nestjs-api"
		} else if allDeps["express"] {
			framework = "node-express"
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

		return &DetectedProject{
			Type:        ProjectTypePython,
			Framework:   framework,
			TargetDir:   targetDir,
			ModuleName:  Slugify(baseName),
			ProjectName: baseName,
		}, nil
	}

	return nil, fmt.Errorf("no supported project found in '%s' (must contain go.mod, package.json, or requirements.txt/pyproject.toml)", targetDir)
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

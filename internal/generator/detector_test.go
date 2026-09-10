package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject_GoFiber(t *testing.T) {
	tempDir := t.TempDir()
	goModContent := `module my-fiber-app

go 1.22

require github.com/gofiber/fiber/v2 v2.52.5
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	proj, err := DetectProject(tempDir)
	if err != nil {
		t.Fatalf("DetectProject failed: %v", err)
	}

	if proj.Type != ProjectTypeGo {
		t.Errorf("Expected type %v, got %v", ProjectTypeGo, proj.Type)
	}
	if proj.Framework != "go-fiber" {
		t.Errorf("Expected framework go-fiber, got %s", proj.Framework)
	}
	if proj.ModuleName != "my-fiber-app" {
		t.Errorf("Expected moduleName my-fiber-app, got %s", proj.ModuleName)
	}
}

func TestDetectProject_GoCLI(t *testing.T) {
	tempDir := t.TempDir()
	goModContent := `module my-cli-tool

go 1.23

require (
	github.com/spf13/cobra v1.8.1
	github.com/charmbracelet/bubbletea v1.2.4
)
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	proj, err := DetectProject(tempDir)
	if err != nil {
		t.Fatalf("DetectProject failed: %v", err)
	}

	if proj.Type != ProjectTypeGo {
		t.Errorf("Expected type %v, got %v", ProjectTypeGo, proj.Type)
	}
	if proj.Framework != "go-cli" {
		t.Errorf("Expected framework go-cli, got %s", proj.Framework)
	}
	if proj.ModuleName != "my-cli-tool" {
		t.Errorf("Expected moduleName my-cli-tool, got %s", proj.ModuleName)
	}
}

func TestDetectProject_GoEcho(t *testing.T) {
	tempDir := t.TempDir()
	goModContent := `module echo-service

go 1.22

require github.com/labstack/echo/v4 v4.12.0
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	proj, err := DetectProject(tempDir)
	if err != nil {
		t.Fatalf("DetectProject failed: %v", err)
	}

	if proj.Framework != "go-echo" {
		t.Errorf("Expected framework go-echo, got %s", proj.Framework)
	}
}

func TestDetectProject_NodeFrameworks(t *testing.T) {
	tests := []struct {
		name      string
		pkgJSON   string
		expectedF string
	}{
		{
			name:      "Express",
			pkgJSON:   `{"name": "my-express", "dependencies": {"express": "^4.19.0"}}`,
			expectedF: "node-express",
		},
		{
			name:      "Fastify",
			pkgJSON:   `{"name": "my-fastify", "dependencies": {"fastify": "^4.28.0"}}`,
			expectedF: "fastify-api",
		},
		{
			name:      "Hono",
			pkgJSON:   `{"name": "my-hono", "dependencies": {"hono": "^4.4.0"}}`,
			expectedF: "hono-api",
		},
		{
			name:      "NestJS",
			pkgJSON:   `{"name": "my-nest", "dependencies": {"@nestjs/core": "^10.0.0"}}`,
			expectedF: "nestjs-api",
		},
		{
			name:      "Elysia",
			pkgJSON:   `{"name": "my-elysia", "dependencies": {"elysia": "^1.1.25"}}`,
			expectedF: "bun-elysia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(tt.pkgJSON), 0644); err != nil {
				t.Fatalf("failed to write package.json: %v", err)
			}

			proj, err := DetectProject(tempDir)
			if err != nil {
				t.Fatalf("DetectProject failed: %v", err)
			}

			if proj.Type != ProjectTypeNode {
				t.Errorf("Expected type node, got %s", proj.Type)
			}
			if proj.Framework != tt.expectedF {
				t.Errorf("Expected framework %s, got %s", tt.expectedF, proj.Framework)
			}
		})
	}
}

func TestDetectProject_PythonFastAPI(t *testing.T) {
	tempDir := t.TempDir()
	reqContent := "fastapi>=0.111.0\nuvicorn>=0.30.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte(reqContent), 0644); err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	proj, err := DetectProject(tempDir)
	if err != nil {
		t.Fatalf("DetectProject failed: %v", err)
	}

	if proj.Type != ProjectTypePython {
		t.Errorf("Expected type python, got %s", proj.Type)
	}
	if proj.Framework != "python-fastapi" {
		t.Errorf("Expected framework python-fastapi, got %s", proj.Framework)
	}
}

func TestDetectProject_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	_, err := DetectProject(tempDir)
	if err == nil {
		t.Errorf("Expected error for empty directory, got nil")
	}
}

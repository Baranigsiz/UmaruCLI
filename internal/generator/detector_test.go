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
		{
			name:      "React",
			pkgJSON:   `{"name": "my-react-app", "dependencies": {"react": "^18.3.1"}}`,
			expectedF: "react-vite-ts",
		},
		{
			name:      "Vue",
			pkgJSON:   `{"name": "my-vue-app", "dependencies": {"vue": "^3.4.0"}}`,
			expectedF: "vue-vite-ts",
		},
		{
			name:      "Svelte",
			pkgJSON:   `{"name": "my-svelte-app", "dependencies": {"svelte": "^5.0.0"}}`,
			expectedF: "svelte-vite-ts",
		},
		{
			name:      "NextJS",
			pkgJSON:   `{"name": "my-next-app", "dependencies": {"next": "^14.2.0"}}`,
			expectedF: "nextjs-tailwind",
		},
		{
			name:      "Astro",
			pkgJSON:   `{"name": "my-astro-app", "dependencies": {"astro": "^4.10.0"}}`,
			expectedF: "astro-tailwind",
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

func TestDetectProject_Monorepos(t *testing.T) {
	// 1. Fullstack Go + React
	t.Run("GoReact", func(t *testing.T) {
		tempDir := t.TempDir()
		apiDir := filepath.Join(tempDir, "apps", "api")
		webDir := filepath.Join(tempDir, "apps", "web")
		_ = os.MkdirAll(apiDir, 0755)
		_ = os.MkdirAll(webDir, 0755)
		_ = os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte("module my-api\ngo 1.24\n"), 0644)
		_ = os.WriteFile(filepath.Join(webDir, "package.json"), []byte(`{"name": "web"}`), 0644)

		proj, err := DetectProject(tempDir)
		if err != nil {
			t.Fatalf("DetectProject failed: %v", err)
		}
		if proj.Type != ProjectTypeGo {
			t.Errorf("Expected type go for fullstack-go-react, got %s", proj.Type)
		}
		if proj.Framework != "fullstack-go-react" {
			t.Errorf("Expected framework fullstack-go-react, got %s", proj.Framework)
		}
	})

	// 2. Fullstack TS Monorepo
	t.Run("TSMonorepo", func(t *testing.T) {
		tempDir := t.TempDir()
		apiDir := filepath.Join(tempDir, "apps", "api")
		webDir := filepath.Join(tempDir, "apps", "web")
		_ = os.MkdirAll(apiDir, 0755)
		_ = os.MkdirAll(webDir, 0755)
		_ = os.WriteFile(filepath.Join(apiDir, "package.json"), []byte(`{"name": "api"}`), 0644)
		_ = os.WriteFile(filepath.Join(webDir, "package.json"), []byte(`{"name": "web"}`), 0644)

		proj, err := DetectProject(tempDir)
		if err != nil {
			t.Fatalf("DetectProject failed: %v", err)
		}
		if proj.Type != ProjectTypeNode {
			t.Errorf("Expected type node for fullstack-ts-monorepo, got %s", proj.Type)
		}
		if proj.Framework != "fullstack-ts-monorepo" {
			t.Errorf("Expected framework fullstack-ts-monorepo, got %s", proj.Framework)
		}
	})
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

func TestDetectProject_Rust(t *testing.T) {
	tests := []struct {
		name      string
		cargoTOML string
		expectedF string
		pkgName   string
	}{
		{
			name: "Axum",
			cargoTOML: `[package]
name = "my-axum-service"
version = "0.1.0"
edition = "2021"

[dependencies]
axum = "0.7"
tokio = { version = "1", features = ["full"] }
`,
			expectedF: "rust-axum",
			pkgName:   "my-axum-service",
		},
		{
			name: "Actix",
			cargoTOML: `[package]
name = "my-actix-app"
version = "0.1.0"
edition = "2021"

[dependencies]
actix-web = "4"
`,
			expectedF: "rust-actix",
			pkgName:   "my-actix-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tempDir, "Cargo.toml"), []byte(tt.cargoTOML), 0644); err != nil {
				t.Fatalf("failed to write Cargo.toml: %v", err)
			}

			proj, err := DetectProject(tempDir)
			if err != nil {
				t.Fatalf("DetectProject failed: %v", err)
			}

			if proj.Type != ProjectTypeRust {
				t.Errorf("Expected type rust, got %s", proj.Type)
			}
			if proj.Framework != tt.expectedF {
				t.Errorf("Expected framework %s, got %s", tt.expectedF, proj.Framework)
			}
			if proj.ProjectName != tt.pkgName {
				t.Errorf("Expected projectName %s, got %s", tt.pkgName, proj.ProjectName)
			}
		})
	}
}

func TestDetectProject_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	_, err := DetectProject(tempDir)
	if err == nil {
		t.Errorf("Expected error for empty directory, got nil")
	}
}

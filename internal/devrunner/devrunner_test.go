package devrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectDevCommand_Go(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Go project with cmd/api/main.go
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module myapp\ngo 1.24\nrequire github.com/gofiber/fiber/v2 v2.52.0"), 0644)
	apiDir := filepath.Join(tempDir, "cmd", "api")
	_ = os.MkdirAll(apiDir, 0755)
	_ = os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main\nfunc main(){}"), 0644)

	cfg, err := DetectDevCommand(DevOptions{
		TargetDir: tempDir,
		Port:      "8080",
	})
	if err != nil {
		t.Fatalf("DetectDevCommand failed: %v", err)
	}

	if cfg.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", cfg.Language)
	}
	if cfg.CommandLine != "go run cmd/api/main.go" {
		t.Errorf("Expected command 'go run cmd/api/main.go', got '%s'", cfg.CommandLine)
	}
	if len(cfg.Env) == 0 || cfg.Env[0] != "PORT=8080" {
		t.Errorf("Expected PORT=8080 env variable, got %v", cfg.Env)
	}
}

func TestDetectDevCommand_Node(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Node project with package.json and pnpm-lock.yaml
	pkgJSON := `{"name":"test-app","scripts":{"dev":"vite"}}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSON), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "pnpm-lock.yaml"), []byte("lockfileVersion: 9.0"), 0644)

	cfg, err := DetectDevCommand(DevOptions{
		TargetDir: tempDir,
	})
	if err != nil {
		t.Fatalf("DetectDevCommand failed: %v", err)
	}

	if cfg.Language != "node" {
		t.Errorf("Expected language 'node', got '%s'", cfg.Language)
	}
	if cfg.CommandLine != "pnpm dev" {
		t.Errorf("Expected 'pnpm dev', got '%s'", cfg.CommandLine)
	}

	// 2. Test override package manager to bun
	cfgBun, err := DetectDevCommand(DevOptions{
		TargetDir:      tempDir,
		PackageManager: "bun",
	})
	if err != nil {
		t.Fatalf("DetectDevCommand with bun failed: %v", err)
	}
	if cfgBun.CommandLine != "bun run dev" {
		t.Errorf("Expected 'bun run dev', got '%s'", cfgBun.CommandLine)
	}
}

func TestDetectDevCommand_Python(t *testing.T) {
	tempDir := t.TempDir()

	// Python FastAPI project
	appDir := filepath.Join(tempDir, "app")
	_ = os.MkdirAll(appDir, 0755)
	_ = os.WriteFile(filepath.Join(appDir, "main.py"), []byte("from fastapi import FastAPI\napp = FastAPI()"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte("fastapi\nuvicorn"), 0644)

	cfg, err := DetectDevCommand(DevOptions{
		TargetDir: tempDir,
		Port:      "9000",
		Host:      "0.0.0.0",
	})
	if err != nil {
		t.Fatalf("DetectDevCommand failed: %v", err)
	}

	if cfg.Language != "python" {
		t.Errorf("Expected language 'python', got '%s'", cfg.Language)
	}
	if !strings.Contains(cfg.CommandLine, "uvicorn app.main:app --reload") {
		t.Errorf("Expected uvicorn command, got '%s'", cfg.CommandLine)
	}
	if !strings.Contains(cfg.CommandLine, "--port 9000") || !strings.Contains(cfg.CommandLine, "--host 0.0.0.0") {
		t.Errorf("Expected port and host flags, got '%s'", cfg.CommandLine)
	}
}

func TestDetectDevCommand_Rust(t *testing.T) {
	tempDir := t.TempDir()

	// Rust Axum project
	_ = os.WriteFile(filepath.Join(tempDir, "Cargo.toml"), []byte("[package]\nname = \"myrust\"\nversion = \"0.1.0\"\n[dependencies]\naxum = \"0.7\""), 0644)
	srcDir := filepath.Join(tempDir, "src")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte("fn main(){}"), 0644)

	cfg, err := DetectDevCommand(DevOptions{
		TargetDir: tempDir,
	})
	if err != nil {
		t.Fatalf("DetectDevCommand failed: %v", err)
	}

	if cfg.Language != "rust" {
		t.Errorf("Expected language 'rust', got '%s'", cfg.Language)
	}
	if cfg.CommandLine != "cargo run" {
		t.Errorf("Expected 'cargo run', got '%s'", cfg.CommandLine)
	}
}

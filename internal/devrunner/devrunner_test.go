package devrunner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"umaru/internal/config"
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

func TestDetectNodePackageManager(t *testing.T) {
	// 1. Bun lockfile
	tempBun := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempBun, "bun.lock"), []byte(""), 0644)
	if pm := detectNodePackageManager(tempBun, ""); pm != "bun" {
		t.Errorf("Expected bun from bun.lock, got %s", pm)
	}

	tempBunB := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempBunB, "bun.lockb"), []byte(""), 0644)
	if pm := detectNodePackageManager(tempBunB, ""); pm != "bun" {
		t.Errorf("Expected bun from bun.lockb, got %s", pm)
	}

	// 2. Yarn lockfile
	tempYarn := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempYarn, "yarn.lock"), []byte(""), 0644)
	if pm := detectNodePackageManager(tempYarn, ""); pm != "yarn" {
		t.Errorf("Expected yarn from yarn.lock, got %s", pm)
	}

	// 3. NPM package-lock.json
	tempNpm := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempNpm, "package-lock.json"), []byte("{}"), 0644)
	if pm := detectNodePackageManager(tempNpm, ""); pm != "npm" {
		t.Errorf("Expected npm from package-lock.json, got %s", pm)
	}

	// 4. Override takes precedence
	if pm := detectNodePackageManager(tempYarn, "bun"); pm != "bun" {
		t.Errorf("Expected bun from override, got %s", pm)
	}

	// 5. Fallback to default npm
	tempEmpty := t.TempDir()
	config.SetTestConfigDir(tempEmpty)
	defer config.SetTestConfigDir("")
	if pm := detectNodePackageManager(tempEmpty, ""); pm != "npm" {
		t.Errorf("Expected fallback npm, got %s", pm)
	}

	// 6. Fallback to user config
	_, _ = config.SetConfigValue("package-manager", "pnpm")
	if pm := detectNodePackageManager(tempEmpty, ""); pm != "pnpm" {
		t.Errorf("Expected pnpm from user config, got %s", pm)
	}
}

func TestDetectDevCommand_GoVariants(t *testing.T) {
	// 1. cmd/web/main.go
	tempWeb := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempWeb, "go.mod"), []byte("module webapp\ngo 1.24"), 0644)
	webDir := filepath.Join(tempWeb, "cmd", "web")
	_ = os.MkdirAll(webDir, 0755)
	_ = os.WriteFile(filepath.Join(webDir, "main.go"), []byte("package main\nfunc main(){}"), 0644)

	cfgWeb, err := DetectDevCommand(DevOptions{TargetDir: tempWeb})
	if err != nil || cfgWeb.CommandLine != "go run cmd/web/main.go" {
		t.Errorf("Expected 'go run cmd/web/main.go', got command: %s, err: %v", cfgWeb.CommandLine, err)
	}

	// 2. root main.go
	tempRoot := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempRoot, "go.mod"), []byte("module rootapp\ngo 1.24"), 0644)
	_ = os.WriteFile(filepath.Join(tempRoot, "main.go"), []byte("package main\nfunc main(){}"), 0644)

	cfgRoot, err := DetectDevCommand(DevOptions{TargetDir: tempRoot})
	if err != nil || cfgRoot.CommandLine != "go run main.go" {
		t.Errorf("Expected 'go run main.go', got command: %s, err: %v", cfgRoot.CommandLine, err)
	}

	// 3. fallback go run .
	tempPkg := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempPkg, "go.mod"), []byte("module otherapp\ngo 1.24"), 0644)
	_ = os.WriteFile(filepath.Join(tempPkg, "app.go"), []byte("package otherapp"), 0644)

	cfgPkg, err := DetectDevCommand(DevOptions{TargetDir: tempPkg})
	if err != nil || cfgPkg.CommandLine != "go run ." {
		t.Errorf("Expected 'go run .', got command: %s, err: %v", cfgPkg.CommandLine, err)
	}
}

func TestDetectDevCommand_NodeVariants(t *testing.T) {
	tempDir := t.TempDir()

	// 1. NestJS start:dev script
	pkgJSON := `{"name":"nest-app","scripts":{"start:dev":"nest start --watch"}}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSON), 0644)

	cfg, err := DetectDevCommand(DevOptions{
		TargetDir:      tempDir,
		PackageManager: "pnpm",
	})
	if err != nil || cfg.CommandLine != "pnpm start:dev" {
		t.Errorf("Expected 'pnpm start:dev', got %s, err: %v", cfg.CommandLine, err)
	}

	// 2. start script with npm
	pkgJSONStart := `{"name":"node-app","scripts":{"start":"node index.js"}}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSONStart), 0644)

	cfgStart, err := DetectDevCommand(DevOptions{
		TargetDir:      tempDir,
		PackageManager: "npm",
	})
	if err != nil || cfgStart.CommandLine != "npm start" {
		t.Errorf("Expected 'npm start', got %s, err: %v", cfgStart.CommandLine, err)
	}

	// 3. start script with bun
	cfgBunStart, err := DetectDevCommand(DevOptions{
		TargetDir:      tempDir,
		PackageManager: "bun",
	})
	if err != nil || cfgBunStart.CommandLine != "bun start" {
		t.Errorf("Expected 'bun start', got %s, err: %v", cfgBunStart.CommandLine, err)
	}

	// 4. dev script with yarn
	pkgJSONDev := `{"name":"node-app","scripts":{"dev":"vite"}}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSONDev), 0644)

	cfgYarn, err := DetectDevCommand(DevOptions{
		TargetDir:      tempDir,
		PackageManager: "yarn",
	})
	if err != nil || cfgYarn.CommandLine != "yarn dev" {
		t.Errorf("Expected 'yarn dev', got %s, err: %v", cfgYarn.CommandLine, err)
	}
}

func TestDetectDevCommand_PythonVariants(t *testing.T) {
	tempDir := t.TempDir()

	// root main.py without app/ dir
	_ = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte("requests"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "main.py"), []byte("print('hello')"), 0644)

	cfg, err := DetectDevCommand(DevOptions{TargetDir: tempDir})
	if err != nil || cfg.CommandLine != "python main.py" {
		t.Errorf("Expected 'python main.py', got %s, err: %v", cfg.CommandLine, err)
	}

	// fallback python -m app.main
	tempFallback := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempFallback, "requirements.txt"), []byte("flask"), 0644)
	_ = os.WriteFile(filepath.Join(tempFallback, "setup.py"), []byte(""), 0644)

	cfgFb, err := DetectDevCommand(DevOptions{TargetDir: tempFallback})
	if err != nil || cfgFb.CommandLine != "python -m app.main" {
		t.Errorf("Expected 'python -m app.main', got %s, err: %v", cfgFb.CommandLine, err)
	}
}

func TestDetectDevCommand_Monorepo(t *testing.T) {
	tempDir := t.TempDir()

	// Fullstack Go + React monorepo structure: apps/api and apps/web
	_ = os.MkdirAll(filepath.Join(tempDir, "apps", "api"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "apps", "api", "go.mod"), []byte("module mymonorepo\ngo 1.24"), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, "apps", "web"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "apps", "web", "package.json"), []byte("{}"), 0644)

	cfg, err := DetectDevCommand(DevOptions{TargetDir: tempDir})
	if err != nil {
		t.Fatalf("DetectDevCommand failed for monorepo: %v", err)
	}
	if cfg.CommandLine != "docker-compose up --build" {
		t.Errorf("Expected 'docker-compose up --build', got '%s'", cfg.CommandLine)
	}
}

func TestDetectDevCommand_Errors(t *testing.T) {
	// 1. Non-existent dir
	_, err := DetectDevCommand(DevOptions{TargetDir: "non_existent_folder_xyz_987"})
	if err == nil {
		t.Errorf("Expected error for non-existent directory, got nil")
	}

	// 2. File instead of directory
	tempFile := filepath.Join(t.TempDir(), "somefile.txt")
	_ = os.WriteFile(tempFile, []byte("test"), 0644)
	_, err = DetectDevCommand(DevOptions{TargetDir: tempFile})
	if err == nil {
		t.Errorf("Expected error for file path, got nil")
	}

	// 3. Unknown empty directory
	emptyDir := t.TempDir()
	_, err = DetectDevCommand(DevOptions{TargetDir: emptyDir})
	if err == nil {
		t.Errorf("Expected error for empty directory with unknown structure, got nil")
	}
}

func TestRun(t *testing.T) {
	// 1. Empty command error
	err := Run(context.Background(), &DevConfig{Command: []string{}})
	if err == nil {
		t.Errorf("Expected error for empty command, got nil")
	}

	// 2. Successful execution of standard command (e.g. go version)
	cfg := &DevConfig{
		TargetDir: t.TempDir(),
		Command:   []string{"go", "version"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = Run(ctx, cfg)
	if err != nil {
		t.Errorf("Expected Run(go version) to succeed, got %v", err)
	}
}


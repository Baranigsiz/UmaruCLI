package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectTestCommand_Go(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module myapp\ngo 1.24"), 0644)

	// 1. Default Go test
	cfg, err := DetectTestCommand(TestOptions{TargetDir: tempDir})
	if err != nil {
		t.Fatalf("DetectTestCommand(Go) failed: %v", err)
	}

	if cfg.Language != "go" {
		t.Errorf("expected language go, got %s", cfg.Language)
	}
	if cfg.CommandLine != "go test ./..." {
		t.Errorf("expected 'go test ./...', got %q", cfg.CommandLine)
	}

	// 2. Go test with verbose, race, coverage, filter
	cfgWithFlags, err := DetectTestCommand(TestOptions{
		TargetDir: tempDir,
		Verbose:   true,
		Race:      true,
		Coverage:  true,
		Filter:    "TestUser",
	})
	if err != nil {
		t.Fatalf("DetectTestCommand(Go with flags) failed: %v", err)
	}

	expectedParts := []string{"go", "test", "-v", "-race", "-coverprofile=coverage.out", "-cover", "-run", "TestUser", "./..."}
	expectedCmd := strings.Join(expectedParts, " ")
	if cfgWithFlags.CommandLine != expectedCmd {
		t.Errorf("expected %q, got %q", expectedCmd, cfgWithFlags.CommandLine)
	}
}

func TestDetectTestCommand_Node_Scripts(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Standard npm test
	pkgJSON := `{
		"name": "node-app",
		"scripts": {
			"test": "vitest run",
			"test:watch": "vitest",
			"test:coverage": "vitest run --coverage"
		}
	}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSON), 0644)

	cfg, err := DetectTestCommand(TestOptions{TargetDir: tempDir})
	if err != nil {
		t.Fatalf("DetectTestCommand(Node) failed: %v", err)
	}
	if cfg.CommandLine != "npm test" {
		t.Errorf("expected 'npm test', got %q", cfg.CommandLine)
	}

	// 2. Watch mode
	cfgWatch, err := DetectTestCommand(TestOptions{TargetDir: tempDir, Watch: true})
	if err != nil {
		t.Fatalf("DetectTestCommand(Node watch) failed: %v", err)
	}
	if cfgWatch.CommandLine != "npm run test:watch" {
		t.Errorf("expected 'npm run test:watch', got %q", cfgWatch.CommandLine)
	}

	// 3. Coverage mode
	cfgCov, err := DetectTestCommand(TestOptions{TargetDir: tempDir, Coverage: true})
	if err != nil {
		t.Fatalf("DetectTestCommand(Node coverage) failed: %v", err)
	}
	if cfgCov.CommandLine != "npm run test:coverage" {
		t.Errorf("expected 'npm run test:coverage', got %q", cfgCov.CommandLine)
	}

	// 4. Filter mode
	cfgFilter, err := DetectTestCommand(TestOptions{TargetDir: tempDir, Filter: "Auth"})
	if err != nil {
		t.Fatalf("DetectTestCommand(Node filter) failed: %v", err)
	}
	if cfgFilter.CommandLine != "npm test -- -t Auth" {
		t.Errorf("expected 'npm test -- -t Auth', got %q", cfgFilter.CommandLine)
	}
}

func TestDetectTestCommand_Node_PackageManagers(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name":"app","scripts":{"test":"jest"}}`), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "pnpm-lock.yaml"), []byte("lockfileVersion: '9.0'"), 0644)

	cfg, err := DetectTestCommand(TestOptions{TargetDir: tempDir})
	if err != nil {
		t.Fatalf("DetectTestCommand(pnpm) failed: %v", err)
	}
	if cfg.CommandLine != "pnpm test" {
		t.Errorf("expected 'pnpm test', got %q", cfg.CommandLine)
	}

	// Override with yarn
	cfgYarn, err := DetectTestCommand(TestOptions{TargetDir: tempDir, PackageManager: "yarn"})
	if err != nil {
		t.Fatalf("DetectTestCommand(yarn override) failed: %v", err)
	}
	if cfgYarn.CommandLine != "yarn test" {
		t.Errorf("expected 'yarn test', got %q", cfgYarn.CommandLine)
	}
}

func TestDetectTestCommand_Node_Bun(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name":"bun-app"}`), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "bun.lockb"), []byte(""), 0644)

	cfg, err := DetectTestCommand(TestOptions{
		TargetDir: tempDir,
		Watch:     true,
		Coverage:  true,
		Filter:    "user.test.ts",
	})
	if err != nil {
		t.Fatalf("DetectTestCommand(bun) failed: %v", err)
	}
	expected := "bun test --watch --coverage user.test.ts"
	if cfg.CommandLine != expected {
		t.Errorf("expected %q, got %q", expected, cfg.CommandLine)
	}
}

func TestDetectTestCommand_Python(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte("fastapi\npytest\n"), 0644)

	cfg, err := DetectTestCommand(TestOptions{
		TargetDir: tempDir,
		Verbose:   true,
		Filter:    "test_login",
	})
	if err != nil {
		t.Fatalf("DetectTestCommand(Python) failed: %v", err)
	}

	if cfg.Language != "python" {
		t.Errorf("expected language python, got %s", cfg.Language)
	}
	if !strings.Contains(cfg.CommandLine, "test_login") {
		t.Errorf("expected filter test_login in command: %s", cfg.CommandLine)
	}
}

func TestDetectTestCommand_Rust(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "Cargo.toml"), []byte("[package]\nname = \"demo\"\n"), 0644)

	cfg, err := DetectTestCommand(TestOptions{
		TargetDir: tempDir,
		Verbose:   true,
		Filter:    "test_api",
	})
	if err != nil {
		t.Fatalf("DetectTestCommand(Rust) failed: %v", err)
	}

	if cfg.Language != "rust" {
		t.Errorf("expected language rust, got %s", cfg.Language)
	}
	if cfg.CommandLine != "cargo test test_api -- --nocapture" {
		t.Errorf("expected 'cargo test test_api -- --nocapture', got %q", cfg.CommandLine)
	}
}

func TestDetectTestCommand_InvalidDir(t *testing.T) {
	_, err := DetectTestCommand(TestOptions{TargetDir: "non_existent_folder_xyz_123"})
	if err == nil {
		t.Errorf("expected error for non-existent directory, got nil")
	}
}

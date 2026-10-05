package deploy

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDeployment_Fly(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)

	res, err := GenerateDeployment(DeployOptions{
		TargetDir: tempDir,
		Platform:  PlatformFly,
		AppName:   "custom-fly-app",
		Port:      9000,
	})
	if err != nil {
		t.Fatalf("GenerateDeployment(Fly) failed: %v", err)
	}

	if res.Platform != PlatformFly {
		t.Errorf("expected platform fly, got %s", res.Platform)
	}
	if res.AppName != "custom-fly-app" {
		t.Errorf("expected app name custom-fly-app, got %s", res.AppName)
	}
	if res.Port != 9000 {
		t.Errorf("expected port 9000, got %d", res.Port)
	}

	flyToml := filepath.Join(tempDir, "fly.toml")
	content, err := os.ReadFile(flyToml)
	if err != nil {
		t.Fatalf("fly.toml not found: %v", err)
	}
	if !strings.Contains(string(content), "app = \"custom-fly-app\"") {
		t.Errorf("fly.toml missing custom app name: %s", string(content))
	}
	if !strings.Contains(string(content), "internal_port = 9000") {
		t.Errorf("fly.toml missing port 9000: %s", string(content))
	}

	// Verify ToJSON
	jsonData, err := res.ToJSON(true)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(string(jsonData), "custom-fly-app") {
		t.Errorf("JSON output missing app name")
	}

	// Verify RenderResultTo
	var buf bytes.Buffer
	RenderResultTo(&buf, res, false)
	if !strings.Contains(buf.String(), "UMARU DEPLOY") {
		t.Errorf("RenderResultTo missing header")
	}
}

func TestGenerateDeployment_Railway(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name":"test-node"}`), 0644)

	res, err := GenerateDeployment(DeployOptions{
		TargetDir: tempDir,
		Platform:  PlatformRailway,
	})
	if err != nil {
		t.Fatalf("GenerateDeployment(Railway) failed: %v", err)
	}

	railwayJSON := filepath.Join(tempDir, "railway.json")
	content, err := os.ReadFile(railwayJSON)
	if err != nil {
		t.Fatalf("railway.json not found: %v", err)
	}
	if !strings.Contains(string(content), "railway.schema.json") {
		t.Errorf("railway.json missing schema: %s", string(content))
	}
	if len(res.Instructions) == 0 {
		t.Errorf("expected railway instructions")
	}
}

func TestGenerateDeployment_Render(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte("fastapi\n"), 0644)

	res, err := GenerateDeployment(DeployOptions{
		TargetDir: tempDir,
		Platform:  PlatformRender,
		Port:      8000,
	})
	if err != nil {
		t.Fatalf("GenerateDeployment(Render) failed: %v", err)
	}
	if res.Platform != PlatformRender {
		t.Errorf("expected platform render, got %s", res.Platform)
	}

	renderYaml := filepath.Join(tempDir, "render.yaml")
	content, err := os.ReadFile(renderYaml)
	if err != nil {
		t.Fatalf("render.yaml not found: %v", err)
	}
	if !strings.Contains(string(content), "runtime: python") {
		t.Errorf("render.yaml missing python runtime: %s", string(content))
	}
}

func TestGenerateDeployment_Docker(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)

	res, err := GenerateDeployment(DeployOptions{
		TargetDir: tempDir,
		Platform:  PlatformDocker,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("GenerateDeployment(Docker) failed: %v", err)
	}

	// In dry-run, file should not be on disk
	dockerfile := filepath.Join(tempDir, "Dockerfile.prod")
	if _, err := os.Stat(dockerfile); err == nil {
		t.Errorf("Dockerfile.prod should NOT exist on disk in dry-run mode")
	}

	if res.Files[0].Action != "create (dry-run)" {
		t.Errorf("expected create (dry-run) action, got %s", res.Files[0].Action)
	}
}

func TestGenerateDeployment_UnsupportedPlatform(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)

	_, err := GenerateDeployment(DeployOptions{
		TargetDir: tempDir,
		Platform:  "invalid-cloud",
	})
	if err == nil {
		t.Fatalf("expected error for unsupported platform")
	}
}

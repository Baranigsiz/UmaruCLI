package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuditProjectAddons_CleanProject(t *testing.T) {
	tempDir := t.TempDir()

	// Simulate Go project
	goModContent := "module test-audit-clean\n\ngo 1.24\n\nrequire github.com/gofiber/fiber/v2 v2.52.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	audit, err := AuditProjectAddons(tempDir)
	if err != nil {
		t.Fatalf("AuditProjectAddons failed: %v", err)
	}

	if audit.Framework != "go-fiber" {
		t.Errorf("Expected framework 'go-fiber', got '%s'", audit.Framework)
	}

	for _, a := range audit.Addons {
		if a.Installed {
			t.Errorf("Expected addon %s to not be installed in clean project, but it was", a.ID)
		}
	}
}

func TestAuditProjectAddons_WithAddons(t *testing.T) {
	tempDir := t.TempDir()

	// Scaffold go-fiber project with addons
	cfg := ProjectConfig{
		ProjectName: "test-audit-installed",
		SafeName:    "test-audit-installed",
		ModuleName:  "test-audit-installed",
		TargetDir:   tempDir,
		Template:    "go-fiber",
		Addons: AddonConfig{
			Docker:   true,
			Database: "postgres",
			Redis:    true,
		},
	}

	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if err := GenerateAddons(cfg); err != nil {
		t.Fatalf("GenerateAddons failed: %v", err)
	}

	audit, err := AuditProjectAddons(tempDir)
	if err != nil {
		t.Fatalf("AuditProjectAddons failed: %v", err)
	}

	expectedInstalled := map[string]bool{
		"docker":   true,
		"postgres": true,
		"redis":    true,
	}

	for _, a := range audit.Addons {
		expected := expectedInstalled[a.ID]
		if a.Installed != expected {
			t.Errorf("Addon %s installed state: got %v, want %v", a.ID, a.Installed, expected)
		}
		if expected && len(a.DetectedFiles) == 0 {
			t.Errorf("Expected detected files for installed addon %s, got none", a.ID)
		}
	}
}

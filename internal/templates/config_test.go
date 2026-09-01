package templates

import (
	"testing"
)

func TestGetAvailableTemplates(t *testing.T) {
	templates, err := GetAvailableTemplates()
	if err != nil {
		t.Fatalf("GetAvailableTemplates() failed: %v", err)
	}

	if len(templates) == 0 {
		t.Fatalf("Expected at least 1 template, got 0")
	}

	for _, tmpl := range templates {
		if tmpl.ID == "" {
			t.Errorf("Template has empty ID: %+v", tmpl)
		}
		if tmpl.Name == "" {
			t.Errorf("Template %s has empty Name", tmpl.ID)
		}
		if tmpl.Description == "" {
			t.Errorf("Template %s has empty Description", tmpl.ID)
		}
	}
}

func TestTemplateConfig_NodeHelpers(t *testing.T) {
	nodeTmpl := TemplateConfig{
		ID:             "react-vite-ts",
		InstallCommand: []string{"npm", "install"},
		RunCommand:     "npm run dev",
	}

	if !nodeTmpl.IsNodeBased() {
		t.Errorf("Expected react-vite-ts to be node-based")
	}

	// Test pnpm
	pnpmInstall := nodeTmpl.GetInstallCommand("pnpm")
	if len(pnpmInstall) != 2 || pnpmInstall[0] != "pnpm" || pnpmInstall[1] != "install" {
		t.Errorf("Expected pnpm install, got %v", pnpmInstall)
	}
	pnpmRun := nodeTmpl.GetRunCommand("pnpm")
	if pnpmRun != "pnpm dev" {
		t.Errorf("Expected 'pnpm dev', got '%s'", pnpmRun)
	}

	// Test bun
	bunInstall := nodeTmpl.GetInstallCommand("bun")
	if len(bunInstall) != 2 || bunInstall[0] != "bun" || bunInstall[1] != "install" {
		t.Errorf("Expected bun install, got %v", bunInstall)
	}
	bunRun := nodeTmpl.GetRunCommand("bun")
	if bunRun != "bun run dev" {
		t.Errorf("Expected 'bun run dev', got '%s'", bunRun)
	}

	// Test non-node template
	goTmpl := TemplateConfig{
		ID:             "go-fiber",
		InstallCommand: []string{"go", "mod", "tidy"},
		RunCommand:     "go run cmd/api/main.go",
	}
	if goTmpl.IsNodeBased() {
		t.Errorf("Expected go-fiber not to be node-based")
	}
	if goTmpl.GetInstallCommand("pnpm")[0] != "go" {
		t.Errorf("Expected go template to preserve install command")
	}
}

func TestFindTemplateByID(t *testing.T) {
	// 1. Existing template
	tmpl, err := FindTemplateByID("go-fiber")
	if err != nil {
		t.Fatalf("Expected to find 'go-fiber', got error: %v", err)
	}
	if tmpl.ID != "go-fiber" {
		t.Errorf("Expected ID 'go-fiber', got '%s'", tmpl.ID)
	}

	// 2. Non-existent template
	_, err = FindTemplateByID("non-existent-template-xyz")
	if err == nil {
		t.Errorf("Expected error for non-existent template, got nil")
	}
}




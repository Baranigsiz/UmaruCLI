package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"umaru/internal/templates"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Cool App", "my-cool-app"},
		{"Hello_World-123", "hello_world-123"},
		{"   Spaces and special @#$ characters   ", "spaces-and-special-characters"},
		{"", "umaru-app"},
		{"---", "umaru-app"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestResolveProjectConfig(t *testing.T) {
	// Case 1: Standard name
	cfg, err := ResolveProjectConfig("my-api", "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	if cfg.ProjectName != "my-api" || cfg.SafeName != "my-api" || cfg.TargetDir != "my-api" {
		t.Errorf("Unexpected config: %+v", cfg)
	}

	// Case 2: Current directory '.'
	cfgDot, err := ResolveProjectConfig(".", "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig('.') failed: %v", err)
	}
	if cfgDot.TargetDir != "." || cfgDot.SafeName == "" || cfgDot.SafeName == "." {
		t.Errorf("Unexpected config for '.': %+v", cfgDot)
	}

	// Case 3: Spaces and uppercase
	cfgSpaces, err := ResolveProjectConfig("My Awesome Backend", "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	if cfgSpaces.SafeName != "my-awesome-backend" {
		t.Errorf("Expected SafeName 'my-awesome-backend', got %s", cfgSpaces.SafeName)
	}
}

func TestGenerate_GoFiber(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "my-fiber-project")

	config, err := ResolveProjectConfig(targetPath, "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}

	// Run the generator
	err = Generate(config)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify expected files are generated without .tmpl extension
	expectedFiles := []string{
		"go.mod",
		".gitignore",
		"Dockerfile",
		"docker-compose.yml",
		"Makefile",
		filepath.Join("cmd", "api", "main.go"),
		filepath.Join("internal", "config", "config.go"),
		filepath.Join("internal", "handlers", "health.go"),
		filepath.Join("internal", "routes", "routes.go"),
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(targetPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not generated", filePath)
		}
	}

	// Verify template.json is EXCLUDED from generated project
	templateJSONPath := filepath.Join(targetPath, "template.json")
	if _, err := os.Stat(templateJSONPath); !os.IsNotExist(err) {
		t.Errorf("template.json was copied to destination, but should have been excluded")
	}

	// Verify template rendering worked on go.mod
	goModBytes, err := os.ReadFile(filepath.Join(targetPath, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read generated go.mod: %v", err)
	}

	goModContent := string(goModBytes)
	if !strings.Contains(goModContent, "module my-fiber-project") {
		t.Errorf("go.mod does not contain expected module definition: %s", goModContent)
	}
}

func TestGenerate_AllTemplates(t *testing.T) {
	availableTemplates, err := templates.GetAvailableTemplates()
	if err != nil {
		t.Fatalf("GetAvailableTemplates failed: %v", err)
	}

	for _, tmpl := range availableTemplates {
		t.Run("Template_"+tmpl.ID, func(t *testing.T) {
			tempDir := t.TempDir()
			targetPath := filepath.Join(tempDir, "sample-project")

			config, err := ResolveProjectConfig(targetPath, tmpl.ID)
			if err != nil {
				t.Fatalf("ResolveProjectConfig failed for %s: %v", tmpl.ID, err)
			}

			err = Generate(config)
			if err != nil {
				t.Fatalf("Generate failed for template %s: %v", tmpl.ID, err)
			}

			// Verify template.json is never included
			templateJSONPath := filepath.Join(targetPath, "template.json")
			if _, err := os.Stat(templateJSONPath); !os.IsNotExist(err) {
				t.Errorf("template.json was found in scaffolded output for %s", tmpl.ID)
			}
		})
	}
}

func TestCheckDestination(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Non-existent path -> should pass
	nonExistent := filepath.Join(tempDir, "brand-new-project")
	if err := CheckDestination(nonExistent, false); err != nil {
		t.Errorf("Expected non-existent path to pass, got error: %v", err)
	}

	// 2. Empty directory -> should pass
	emptyDir := filepath.Join(tempDir, "empty-dir")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}
	if err := CheckDestination(emptyDir, false); err != nil {
		t.Errorf("Expected empty dir to pass, got error: %v", err)
	}

	// 3. Non-empty directory without force -> should fail
	dummyFile := filepath.Join(emptyDir, "some-file.txt")
	if err := os.WriteFile(dummyFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	if err := CheckDestination(emptyDir, false); err == nil {
		t.Errorf("Expected non-empty directory without force to fail, but it passed")
	}

	// 4. Non-empty directory with force -> should pass
	if err := CheckDestination(emptyDir, true); err != nil {
		t.Errorf("Expected non-empty directory with force=true to pass, got error: %v", err)
	}

	// 5. Existing file (not a directory) -> should fail
	if err := CheckDestination(dummyFile, false); err == nil {
		t.Errorf("Expected file path to fail, but it passed")
	}
}

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "my-test-project")

	config := ProjectConfig{
		ProjectName: targetPath,
		Template:    "go-fiber",
	}

	// Run the generator
	err := Generate(config)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Verify expected files are generated without .tmpl extension
	expectedFiles := []string{
		"go.mod",
		"main.go",
		".gitignore",
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
	if !strings.Contains(goModContent, "module") {
		t.Errorf("go.mod does not contain expected module definition: %s", goModContent)
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

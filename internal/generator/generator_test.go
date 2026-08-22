package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	// Use t.TempDir() to create a temporary directory for output
	// This prevents the test from polluting the real filesystem
	tempDir := t.TempDir()

	// The target project name/path should be inside the temp dir
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

	// Verify that the files were generated correctly
	// We expect go-fiber to generate at least go.mod, main.go, and template.json
	
	expectedFiles := []string{
		"go.mod",
		"main.go",
		".gitignore", // Since we strip .tmpl, it should be .gitignore, not .gitignore.tmpl
		"template.json",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(targetPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not generated", filePath)
		}
	}

	// Additionally, verify that template rendering worked by checking the content of go.mod
	// In go.mod.tmpl, the module name is likely {{.ProjectName}}, but we passed targetPath as ProjectName.
	// Since targetPath is an absolute path, it will be rendered as such. We just need to check if it's there.
	goModBytes, err := os.ReadFile(filepath.Join(targetPath, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read generated go.mod: %v", err)
	}

	goModContent := string(goModBytes)
	if goModContent == "" {
		t.Errorf("go.mod is empty, template rendering might have failed")
	}
}

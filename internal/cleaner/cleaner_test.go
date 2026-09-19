package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 350, "350.0 MB"},
		{1024 * 1024 * 1024 * 2, "2.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, got, tt.expected)
		}
	}
}

func TestScanAndClean(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create realistic dummy project structure
	// Normal project files (MUST NOT BE DELETED)
	srcDir := filepath.Join(tempDir, "src")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\nfunc main(){}"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte("{}"), 0644)

	// Git folder (MUST NEVER BE TOUCHED)
	gitDir := filepath.Join(tempDir, ".git", "objects")
	_ = os.MkdirAll(gitDir, 0755)
	_ = os.WriteFile(filepath.Join(gitDir, "commit-hash"), []byte("data"), 0644)

	// Removable targets
	nodeModulesDir := filepath.Join(tempDir, "node_modules", "express")
	_ = os.MkdirAll(nodeModulesDir, 0755)
	_ = os.WriteFile(filepath.Join(nodeModulesDir, "index.js"), []byte("module.exports = {}"), 0644)

	distDir := filepath.Join(tempDir, "dist")
	_ = os.MkdirAll(distDir, 0755)
	_ = os.WriteFile(filepath.Join(distDir, "bundle.js"), []byte("console.log(1)"), 0644)

	pycacheDir := filepath.Join(tempDir, "__pycache__")
	_ = os.MkdirAll(pycacheDir, 0755)
	_ = os.WriteFile(filepath.Join(pycacheDir, "test.cpython-312.pyc"), []byte("pyc"), 0644)

	venvDir := filepath.Join(tempDir, ".venv")
	_ = os.MkdirAll(venvDir, 0755)
	_ = os.WriteFile(filepath.Join(venvDir, "pyvenv.cfg"), []byte("cfg"), 0644)

	// 2. Scan without IncludeAll (should NOT include .venv)
	report, err := Scan(ScanOptions{
		RootDir:    tempDir,
		Recursive:  true,
		IncludeAll: false,
	})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(report.Items) == 0 {
		t.Fatalf("Expected removable items in scan, got 0")
	}

	foundNode := false
	foundDist := false
	foundPycache := false
	foundVenv := false

	for _, item := range report.Items {
		switch item.Name {
		case "node_modules":
			foundNode = true
		case "dist":
			foundDist = true
		case "__pycache__":
			foundPycache = true
		case ".venv":
			foundVenv = true
		}
	}

	if !foundNode || !foundDist || !foundPycache {
		t.Errorf("Expected node_modules, dist, and __pycache__ to be found. Found: %+v", report.Items)
	}
	if foundVenv {
		t.Errorf(".venv should NOT be found when IncludeAll is false")
	}
	if report.TotalSizeBytes <= 0 {
		t.Errorf("Expected TotalSizeBytes > 0, got %d", report.TotalSizeBytes)
	}

	// 3. Scan WITH IncludeAll (should include .venv)
	reportAll, err := Scan(ScanOptions{
		RootDir:    tempDir,
		Recursive:  true,
		IncludeAll: true,
	})
	if err != nil {
		t.Fatalf("Scan with IncludeAll failed: %v", err)
	}
	hasVenv := false
	for _, item := range reportAll.Items {
		if item.Name == ".venv" {
			hasVenv = true
			break
		}
	}
	if !hasVenv {
		t.Errorf("Expected .venv to be found when IncludeAll is true")
	}

	// 4. Execute Clean
	result, err := ExecuteClean(report)
	if err != nil {
		t.Fatalf("ExecuteClean failed: %v", err)
	}

	if result.DeletedCount != len(report.Items) {
		t.Errorf("Expected %d items deleted, got %d", len(report.Items), result.DeletedCount)
	}

	// 5. Verify deleted items are gone
	if _, err := os.Stat(nodeModulesDir); !os.IsNotExist(err) {
		t.Errorf("node_modules was not deleted")
	}
	if _, err := os.Stat(distDir); !os.IsNotExist(err) {
		t.Errorf("dist was not deleted")
	}

	// 6. Verify protected files are preserved
	if _, err := os.Stat(filepath.Join(srcDir, "main.go")); os.IsNotExist(err) {
		t.Errorf("CRITICAL: src/main.go was accidentally deleted!")
	}
	if _, err := os.Stat(filepath.Join(tempDir, ".git", "objects", "commit-hash")); os.IsNotExist(err) {
		t.Errorf("CRITICAL: .git directory was accidentally deleted!")
	}
}

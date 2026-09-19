package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateLicenseFile_StandardLicenses(t *testing.T) {
	licenses := []string{
		"MIT",
		"Apache-2.0",
		"BSD-3-Clause",
		"ISC",
		"GPL-3.0",
		"Unlicense",
	}

	for _, lic := range licenses {
		tempDir := t.TempDir()
		author := "Jane Developer"

		err := GenerateLicenseFile(tempDir, lic, author)
		if err != nil {
			t.Fatalf("GenerateLicenseFile(%s) failed: %v", lic, err)
		}

		licPath := filepath.Join(tempDir, "LICENSE")
		data, err := os.ReadFile(licPath)
		if err != nil {
			t.Fatalf("Failed to read generated LICENSE for %s: %v", lic, err)
		}

		content := string(data)
		if len(content) == 0 {
			t.Errorf("Expected non-empty LICENSE for %s", lic)
		}

		if lic != "Unlicense" && !strings.Contains(content, author) {
			t.Errorf("Expected LICENSE for %s to contain author %q, got: %s", lic, author, content)
		}
	}
}

func TestGenerateLicenseFile_NoneOrEmpty(t *testing.T) {
	tempDir := t.TempDir()
	if err := GenerateLicenseFile(tempDir, "none", "Author"); err != nil {
		t.Fatalf("Unexpected error for 'none': %v", err)
	}

	licPath := filepath.Join(tempDir, "LICENSE")
	if _, err := os.Stat(licPath); !os.IsNotExist(err) {
		t.Errorf("Expected LICENSE to NOT exist when license is 'none'")
	}
}

func TestGenerateLicenseFile_Idempotency(t *testing.T) {
	tempDir := t.TempDir()
	customLicense := "Custom Pre-existing License"
	licPath := filepath.Join(tempDir, "LICENSE")
	if err := os.WriteFile(licPath, []byte(customLicense), 0644); err != nil {
		t.Fatalf("Failed to write custom license: %v", err)
	}

	// Generating MIT should NOT overwrite existing LICENSE
	if err := GenerateLicenseFile(tempDir, "MIT", "New Author"); err != nil {
		t.Fatalf("GenerateLicenseFile failed: %v", err)
	}

	data, err := os.ReadFile(licPath)
	if err != nil {
		t.Fatalf("Failed to read LICENSE: %v", err)
	}
	if string(data) != customLicense {
		t.Errorf("Expected existing LICENSE to be preserved, got: %s", string(data))
	}
}

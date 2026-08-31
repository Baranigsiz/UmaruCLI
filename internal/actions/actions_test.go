package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"umaru/internal/templates"
)

func TestInitGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH, skipping TestInitGit")
	}

	tempDir := t.TempDir()
	err := InitGit(tempDir)
	if err != nil {
		t.Fatalf("InitGit() failed: %v", err)
	}

	gitDir := filepath.Join(tempDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("Expected .git directory to exist at %s", gitDir)
	}
}

func TestInstallDependencies_Empty(t *testing.T) {
	tempDir := t.TempDir()
	err := InstallDependencies(tempDir, templates.TemplateConfig{
		InstallCommand: []string{},
	})
	if err != nil {
		t.Errorf("Expected empty install command to succeed, got %v", err)
	}
}

package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
	err := InstallDependencies(tempDir, []string{}, false)
	if err != nil {
		t.Errorf("Expected empty install command to succeed, got %v", err)
	}
}

func TestBuildCommand_Nil(t *testing.T) {
	cmd := buildCommand(".", []string{})
	if cmd != nil {
		t.Errorf("Expected nil cmd for empty slice, got %+v", cmd)
	}
}

func TestCommitGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH, skipping TestCommitGit")
	}

	tempDir := t.TempDir()
	if err := InitGit(tempDir); err != nil {
		t.Fatalf("InitGit() failed: %v", err)
	}

	dummyFile := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(dummyFile, []byte("# Test Project"), 0644); err != nil {
		t.Fatalf("Failed to write dummy file: %v", err)
	}

	err := CommitGit(tempDir, "chore: initial commit")
	if err != nil {
		t.Fatalf("CommitGit() failed: %v", err)
	}
}


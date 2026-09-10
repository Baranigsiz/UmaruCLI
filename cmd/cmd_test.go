package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	os.Stderr = w

	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
	rootCmd.SetArgs(args)

	var captured bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&captured, r)
		close(done)
	}()

	execErr := rootCmd.Execute()

	_ = w.Close()
	<-done
	_ = r.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return captured.String(), execErr
}

func TestVersionCmd(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	if !strings.Contains(out, "Umaru CLI") {
		t.Errorf("Expected version output to contain 'Umaru CLI', got: %s", out)
	}
}

func TestListCmd(t *testing.T) {
	out, err := executeCommand("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	if !strings.Contains(out, "Available Starter Templates") {
		t.Errorf("Expected list output to contain 'Available Starter Templates', got: %s", out)
	}
	if !strings.Contains(out, "go-fiber") {
		t.Errorf("Expected list output to contain 'go-fiber', got: %s", out)
	}
	if !strings.Contains(out, "react-vite-ts") {
		t.Errorf("Expected list output to contain 'react-vite-ts', got: %s", out)
	}
}

func TestInfoCmd(t *testing.T) {
	out, err := executeCommand("info", "go-fiber")
	if err != nil {
		t.Fatalf("info command failed: %v", err)
	}

	if !strings.Contains(out, "go-fiber") {
		t.Errorf("Expected info output to contain 'go-fiber', got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("Expected info output to contain port '8080', got: %s", out)
	}
}

func TestInitCmd_Aliases(t *testing.T) {
	// Verify initCmd has "new" and "create" aliases
	aliases := initCmd.Aliases
	hasNew := false
	hasCreate := false

	for _, a := range aliases {
		if a == "new" {
			hasNew = true
		}
		if a == "create" {
			hasCreate = true
		}
	}

	if !hasNew {
		t.Errorf("Expected initCmd to have 'new' alias")
	}
	if !hasCreate {
		t.Errorf("Expected initCmd to have 'create' alias")
	}

	// Verify command resolution through rootCmd
	cmdNew, _, err := rootCmd.Find([]string{"new"})
	if err != nil || cmdNew.Name() != "init" {
		t.Errorf("Expected 'new' alias to resolve to 'init', got: %v", cmdNew)
	}

	cmdCreate, _, err := rootCmd.Find([]string{"create"})
	if err != nil || cmdCreate.Name() != "init" {
		t.Errorf("Expected 'create' alias to resolve to 'init', got: %v", cmdCreate)
	}
}

func TestConfigCmd_GetAndList(t *testing.T) {
	// Test config list
	outList, err := executeCommand("config", "list")
	if err != nil {
		t.Fatalf("config list failed: %v", err)
	}
	if !strings.Contains(outList, "Global Configuration") {
		t.Errorf("Expected config list output to contain 'Global Configuration', got: %s", outList)
	}

	// Test config get license (default is MIT)
	outGet, err := executeCommand("config", "get", "license")
	if err != nil {
		t.Fatalf("config get license failed: %v", err)
	}
	if !strings.Contains(outGet, "MIT") {
		t.Errorf("Expected config get license to return 'MIT', got: %s", outGet)
	}
}

func TestCompletionCmd(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		out, err := executeCommand("completion", shell)
		if err != nil {
			t.Errorf("completion %s failed: %v", shell, err)
		}
		if len(out) == 0 {
			t.Errorf("Expected non-empty output for completion %s", shell)
		}
	}
}

func TestAddCmd_CI(t *testing.T) {
	tempDir := t.TempDir()
	goMod := "module test-app\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	out, err := executeCommand("add", "ci", "--dir", tempDir)
	if err != nil {
		t.Fatalf("add ci failed: %v", err)
	}

	if !strings.Contains(out, "Addon(s) Injected Successfully") {
		t.Errorf("Expected success output, got: %s", out)
	}
	if !strings.Contains(out, "CI/CD: GitHub Actions") {
		t.Errorf("Expected output to mention CI/CD: GitHub Actions, got: %s", out)
	}

	ciPath := filepath.Join(tempDir, ".github", "workflows", "ci.yml")
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		t.Errorf("Expected .github/workflows/ci.yml to be created, but it doesn't exist")
	}
}


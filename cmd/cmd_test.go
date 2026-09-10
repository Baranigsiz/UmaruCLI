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

func TestListCmd_CategoryFilter(t *testing.T) {
	// Filter by frontend
	out, err := executeCommand("list", "-c", "Frontend")
	if err != nil {
		t.Fatalf("list -c Frontend failed: %v", err)
	}
	if !strings.Contains(out, "react-vite-ts") {
		t.Errorf("Expected frontend list to contain 'react-vite-ts', got: %s", out)
	}
	if strings.Contains(out, "go-fiber") {
		t.Errorf("Expected frontend list to NOT contain 'go-fiber', got: %s", out)
	}

	// Filter by backend
	outBackend, err := executeCommand("list", "--category", "Backend")
	if err != nil {
		t.Fatalf("list --category Backend failed: %v", err)
	}
	if !strings.Contains(outBackend, "go-fiber") {
		t.Errorf("Expected backend list to contain 'go-fiber', got: %s", outBackend)
	}
	if strings.Contains(outBackend, "react-vite-ts") {
		t.Errorf("Expected backend list to NOT contain 'react-vite-ts', got: %s", outBackend)
	}
}

func TestInitCmd_DockerAndCIFlags(t *testing.T) {
	dockerFlag := initCmd.Flags().Lookup("docker")
	if dockerFlag == nil {
		t.Fatalf("Expected initCmd to have --docker flag")
	}
	ciFlag := initCmd.Flags().Lookup("ci")
	if ciFlag == nil {
		t.Fatalf("Expected initCmd to have --ci flag")
	}
	categoryFlag := listCmd.Flags().Lookup("category")
	if categoryFlag == nil {
		t.Fatalf("Expected listCmd to have --category flag")
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

func TestCompletionCmd_Install(t *testing.T) {
	tempHome := t.TempDir()
	completionTestHomeDir = tempHome
	defer func() {
		completionTestHomeDir = ""
	}()

	// 1. Test installation for bash
	outBash, err := executeCommand("completion", "bash", "--install")
	if err != nil {
		t.Fatalf("completion bash --install failed: %v", err)
	}
	if !strings.Contains(outBash, "Successfully installed") {
		t.Errorf("Expected success output, got: %s", outBash)
	}
	bashrcPath := filepath.Join(tempHome, ".bashrc")
	if _, err := os.Stat(bashrcPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", bashrcPath)
	}

	// 1b. Test idempotence for bash
	outBash2, err := executeCommand("completion", "bash", "--install")
	if err != nil {
		t.Fatalf("second completion bash --install failed: %v", err)
	}
	if !strings.Contains(outBash2, "already installed") {
		t.Errorf("Expected idempotency message, got: %s", outBash2)
	}

	// 2. Test installation for zsh
	outZsh, err := executeCommand("completion", "zsh", "--install")
	if err != nil {
		t.Fatalf("completion zsh --install failed: %v", err)
	}
	if !strings.Contains(outZsh, "Successfully installed") {
		t.Errorf("Expected success output for zsh, got: %s", outZsh)
	}
	zshrcPath := filepath.Join(tempHome, ".zshrc")
	if _, err := os.Stat(zshrcPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", zshrcPath)
	}

	// 3. Test installation for powershell
	outPS, err := executeCommand("completion", "powershell", "--install")
	if err != nil {
		t.Fatalf("completion powershell --install failed: %v", err)
	}
	if !strings.Contains(outPS, "Successfully installed") {
		t.Errorf("Expected success output for powershell, got: %s", outPS)
	}

	// 4. Test installation for fish
	outFish, err := executeCommand("completion", "fish", "--install")
	if err != nil {
		t.Fatalf("completion fish --install failed: %v", err)
	}
	if !strings.Contains(outFish, "Successfully installed") {
		t.Errorf("Expected success output for fish, got: %s", outFish)
	}
	fishFile := filepath.Join(tempHome, ".config", "fish", "completions", "umaru.fish")
	if _, err := os.Stat(fishFile); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", fishFile)
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

func TestConfigCmd_Completions(t *testing.T) {
	// Test config get completions
	if configGetCmd.ValidArgsFunction == nil {
		t.Fatalf("Expected configGetCmd to have ValidArgsFunction")
	}
	completions, _ := configGetCmd.ValidArgsFunction(configGetCmd, []string{}, "")
	if len(completions) != 4 {
		t.Errorf("Expected 4 config keys in completion, got %d: %v", len(completions), completions)
	}

	// Test config set completions (key arg)
	if configSetCmd.ValidArgsFunction == nil {
		t.Fatalf("Expected configSetCmd to have ValidArgsFunction")
	}
	setKeyCompletions, _ := configSetCmd.ValidArgsFunction(configSetCmd, []string{}, "")
	if len(setKeyCompletions) != 4 {
		t.Errorf("Expected 4 config keys in set completion, got %d", len(setKeyCompletions))
	}

	// Test config set completions (value arg for package-manager)
	pmCompletions, _ := configSetCmd.ValidArgsFunction(configSetCmd, []string{"package-manager"}, "")
	if len(pmCompletions) != 4 {
		t.Errorf("Expected 4 package managers in completion, got %d: %v", len(pmCompletions), pmCompletions)
	}
}

func TestAddCmd_Completions(t *testing.T) {
	if addCmd.ValidArgsFunction == nil {
		t.Fatalf("Expected addCmd to have ValidArgsFunction")
	}
	completions, _ := addCmd.ValidArgsFunction(addCmd, []string{}, "")
	if len(completions) != 6 {
		t.Errorf("Expected 6 addons in completion, got %d: %v", len(completions), completions)
	}

	// Filtered completions when redis is already selected
	filtered, _ := addCmd.ValidArgsFunction(addCmd, []string{"redis"}, "")
	if len(filtered) != 5 {
		t.Errorf("Expected 5 addons in completion after selecting redis, got %d", len(filtered))
	}
	for _, f := range filtered {
		if strings.HasPrefix(f, "redis\t") {
			t.Errorf("redis should be filtered out from completion suggestions")
		}
	}
}



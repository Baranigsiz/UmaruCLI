package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	// Reset CLI command flags to ensure complete test isolation
	templateFlag = ""
	packageManagerFlag = ""
	fromFlag = ""
	dbFlag = ""
	authFlag = ""
	redisFlag = false
	dockerFlag = false
	ciFlag = false
	noAddonsFlag = false
	noGitFlag = false
	skipInstallFlag = false
	forceFlag = false
	verboseFlag = false
	dryRunFlag = false
	commitFlag = false
	yesFlag = false
	completionInstallFlag = false
	addDirFlag = "."
	addForceFlag = false
	addListFlag = false
	addJSONFlag = false
	addAllFlag = false
	listCategoryFlag = ""
	listSearchFlag = ""
	listJSONFlag = false
	quietFlag = false
	noColorFlag = false

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

func TestVersionFlag(t *testing.T) {
	out, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("--version flag failed: %v", err)
	}

	if !strings.Contains(out, "Umaru CLI") {
		t.Errorf("Expected --version output to contain 'Umaru CLI', got: %s", out)
	}

	outShort, err := executeCommand("-v")
	if err != nil {
		t.Fatalf("-v flag failed: %v", err)
	}

	if !strings.Contains(outShort, "Umaru CLI") {
		t.Errorf("Expected -v output to contain 'Umaru CLI', got: %s", outShort)
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
	if !strings.Contains(out, "3000") {
		t.Errorf("Expected info output to contain port '3000', got: %s", out)
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

func TestConfigCmd_SetAndReset(t *testing.T) {
	// Set author
	outSet, err := executeCommand("config", "set", "author", "Test Author Name")
	if err != nil {
		t.Fatalf("config set author failed: %v", err)
	}
	if !strings.Contains(outSet, "updated successfully") {
		t.Errorf("Expected success message for config set, got: %s", outSet)
	}

	// Reset
	outReset, err := executeCommand("config", "reset")
	if err != nil {
		t.Fatalf("config reset failed: %v", err)
	}
	if !strings.Contains(outReset, "reset to default") {
		t.Errorf("Expected reset message, got: %s", outReset)
	}
}

func TestDoctorCmd(t *testing.T) {
	out, err := executeCommand("doctor")
	if err != nil {
		t.Fatalf("doctor command failed: %v", err)
	}

	if !strings.Contains(out, "UMARU DOCTOR") {
		t.Errorf("Expected doctor output to contain 'UMARU DOCTOR', got: %s", out)
	}
	if !strings.Contains(out, "Template Ecosystem Readiness") {
		t.Errorf("Expected doctor output to contain template readiness, got: %s", out)
	}
}

func TestDoctorCmd_Verbose(t *testing.T) {
	out, err := executeCommand("doctor", "-v")
	if err != nil {
		t.Fatalf("doctor -v command failed: %v", err)
	}

	if !strings.Contains(out, "BINARY PATH") {
		t.Errorf("Expected verbose doctor output to contain 'BINARY PATH', got: %s", out)
	}

	outFull, err := executeCommand("doctor", "--verbose")
	if err != nil {
		t.Fatalf("doctor --verbose command failed: %v", err)
	}
	if !strings.Contains(outFull, "BINARY PATH") {
		t.Errorf("Expected verbose doctor output to contain 'BINARY PATH', got: %s", outFull)
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

func resetInitFlags() {
	dbFlag = ""
	authFlag = ""
	packageManagerFlag = ""
	templateFlag = ""
	fromFlag = ""
	redisFlag = false
	dockerFlag = false
	ciFlag = false
	noAddonsFlag = false
	noGitFlag = false
	skipInstallFlag = false
	forceFlag = false
	verboseFlag = false
	dryRunFlag = false
	commitFlag = false
	yesFlag = false
}

func TestInitCmd_YesFlag(t *testing.T) {
	flag := initCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatalf("Expected initCmd to have --yes flag")
	}
	if flag.Shorthand != "y" {
		t.Errorf("Expected shorthand 'y', got '%s'", flag.Shorthand)
	}
}

func TestInitCmd_FlagValidation(t *testing.T) {
	defer resetInitFlags()

	// 1. Invalid DB driver
	resetInitFlags()
	_, err := executeCommand("init", "--db", "oracle")
	if err == nil {
		t.Fatalf("Expected error for invalid --db, got nil")
	}
	if !strings.Contains(err.Error(), "invalid database driver 'oracle'") {
		t.Errorf("Expected error to mention invalid database driver, got: %v", err)
	}

	// 2. Invalid Auth option
	resetInitFlags()
	_, err = executeCommand("init", "--auth", "oauth")
	if err == nil {
		t.Fatalf("Expected error for invalid --auth, got nil")
	}
	if !strings.Contains(err.Error(), "invalid authentication option 'oauth'") {
		t.Errorf("Expected error to mention invalid authentication option, got: %v", err)
	}

	// 3. Invalid Package Manager
	resetInitFlags()
	_, err = executeCommand("init", "--package-manager", "pip")
	if err == nil {
		t.Fatalf("Expected error for invalid --package-manager, got nil")
	}
	if !strings.Contains(err.Error(), "invalid package manager 'pip'") {
		t.Errorf("Expected error to mention invalid package manager, got: %v", err)
	}
}

func TestInitCmd_YesFlag_NonInteractiveDryRun(t *testing.T) {
	defer resetInitFlags()
	resetInitFlags()

	out, err := executeCommand("init", "test-dryrun-app", "-t", "react-vite-ts", "-y", "--dry-run")
	if err != nil {
		t.Fatalf("Expected init with -y and --dry-run to succeed, got error: %v", err)
	}

	if !strings.Contains(out, "Dry-Run Mode") {
		t.Errorf("Expected dry run output, got: %s", out)
	}
	if !strings.Contains(out, "package.json") {
		t.Errorf("Expected dry run to show package.json, got: %s", out)
	}
}

func TestListCmd_JSON(t *testing.T) {
	out, err := executeCommand("list", "--json")
	if err != nil {
		t.Fatalf("list --json failed: %v", err)
	}

	var tmpls []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &tmpls); err != nil {
		t.Fatalf("list --json did not return valid JSON array: %v\nOutput: %s", err, out)
	}
	if len(tmpls) == 0 {
		t.Errorf("Expected templates array to be non-empty")
	}
}

func TestInfoCmd_JSON(t *testing.T) {
	out, err := executeCommand("info", "react-vite-ts", "--json")
	if err != nil {
		t.Fatalf("info --json failed: %v", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("info --json did not return valid JSON: %v\nOutput: %s", err, out)
	}
	if info["Config"] == nil {
		t.Errorf("Expected info JSON to contain Config field")
	}
}

func TestDoctorCmd_JSON(t *testing.T) {
	out, err := executeCommand("doctor", "--json")
	if err != nil {
		t.Fatalf("doctor --json failed: %v", err)
	}

	var report map[string]interface{}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("doctor --json did not return valid JSON: %v\nOutput: %s", err, out)
	}
	if report["System"] == nil {
		t.Errorf("Expected doctor JSON to contain System info")
	}
}

func TestConfigListCmd_JSON(t *testing.T) {
	out, err := executeCommand("config", "list", "--json")
	if err != nil {
		t.Fatalf("config list --json failed: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(out), &cfg); err != nil {
		t.Fatalf("config list --json did not return valid JSON: %v\nOutput: %s", err, out)
	}
}

func TestGlobalFlags_NoColorAndQuiet(t *testing.T) {
	out, err := executeCommand("--no-color", "--quiet", "version")
	if err != nil {
		t.Fatalf("version with --no-color --quiet failed: %v", err)
	}
	if !strings.Contains(out, "Umaru CLI") {
		t.Errorf("Expected version output to contain 'Umaru CLI', got: %s", out)
	}
}

func TestAddCmd_List_TableAndJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Simulate Go project with a Dockerfile
	goModContent := "module test-add-list-app\n\ngo 1.24\n\nrequire github.com/gofiber/fiber/v2 v2.52.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte("version: '3.8'\n"), 0644); err != nil {
		t.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	// 1. Test table output
	out, err := executeCommand("add", "--list", "--dir", tempDir)
	if err != nil {
		t.Fatalf("add --list failed: %v", err)
	}
	if !strings.Contains(out, "Addon Audit") {
		t.Errorf("Expected output to contain 'Addon Audit', got: %s", out)
	}
	if !strings.Contains(out, "Docker & Compose") {
		t.Errorf("Expected output to contain 'Docker & Compose', got: %s", out)
	}

	// 2. Test JSON output
	jsonOut, err := executeCommand("add", "--list", "--json", "--dir", tempDir)
	if err != nil {
		t.Fatalf("add --list --json failed: %v", err)
	}
	var audit map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOut), &audit); err != nil {
		t.Fatalf("add --list --json did not return valid JSON: %v\nOutput: %s", err, jsonOut)
	}
	if audit["project_name"] == nil || audit["addons"] == nil {
		t.Errorf("Expected audit JSON to contain project_name and addons, got: %v", audit)
	}
}

func TestInfoCmd_InvalidTemplate(t *testing.T) {
	_, err := executeCommand("info", "non-existent-template-999")
	if err == nil {
		t.Errorf("Expected error when inspecting non-existent template, got nil")
	}
}

func TestInitCmd_ValidationErrors(t *testing.T) {
	t.Run("InvalidTemplate", func(t *testing.T) {
		tempDir := filepath.Join(t.TempDir(), "invalid-tpl-app")
		_, err := executeCommand("init", tempDir, "-t", "non-existent-template-xyz", "--skip-install", "--no-git")
		if err == nil {
			t.Errorf("Expected error for invalid template ID, got nil")
		}
	})

	t.Run("ExistingNonEmptyDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tempDir, "existing-file.txt"), []byte("data"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		_, err := executeCommand("init", tempDir, "-t", "go-fiber", "--skip-install", "--no-git")
		if err == nil {
			t.Errorf("Expected error when target directory is not empty, got nil")
		}
	})
}

func TestAddCmd_ValidationErrors(t *testing.T) {
	t.Run("EmptyDirectoryWithoutProject", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := executeCommand("add", "docker", "--dir", tempDir)
		if err == nil {
			t.Errorf("Expected error when running add in non-project directory, got nil")
		}
	})
}

func TestConfigCmd_FullFlow(t *testing.T) {
	// 1. Set key
	setOut, err := executeCommand("config", "set", "author", "Test Engineer")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}
	if !strings.Contains(setOut, "Updated") && !strings.Contains(setOut, "Test Engineer") {
		t.Errorf("Unexpected output from config set: %s", setOut)
	}

	// 2. Get key
	getOut, err := executeCommand("config", "get", "author")
	if err != nil {
		t.Fatalf("config get failed: %v", err)
	}
	if !strings.Contains(getOut, "Test Engineer") {
		t.Errorf("Expected config get author to contain 'Test Engineer', got: %s", getOut)
	}

	// 3. List
	listOut, err := executeCommand("config", "list")
	if err != nil {
		t.Fatalf("config list failed: %v", err)
	}
	if !strings.Contains(listOut, "Test Engineer") {
		t.Errorf("Expected config list to contain 'Test Engineer', got: %s", listOut)
	}

	// 4. Reset
	resetOut, err := executeCommand("config", "reset")
	if err != nil {
		t.Fatalf("config reset failed: %v", err)
	}
	if !strings.Contains(resetOut, "reset") {
		t.Errorf("Unexpected output from config reset: %s", resetOut)
	}
}

func TestListCmd_Search(t *testing.T) {
	// 1. Search by keyword
	outFiber, err := executeCommand("list", "-s", "fiber")
	if err != nil {
		t.Fatalf("list -s fiber failed: %v", err)
	}
	if !strings.Contains(outFiber, "go-fiber") {
		t.Errorf("Expected search 'fiber' to contain 'go-fiber', got: %s", outFiber)
	}
	if strings.Contains(outFiber, "react-vite-ts") {
		t.Errorf("Expected search 'fiber' to NOT contain 'react-vite-ts', got: %s", outFiber)
	}

	// 2. Search by category & search combined
	outCombo, err := executeCommand("list", "-c", "Backend", "--search", "echo")
	if err != nil {
		t.Fatalf("list -c Backend --search echo failed: %v", err)
	}
	if !strings.Contains(outCombo, "go-echo") {
		t.Errorf("Expected combo search to contain 'go-echo', got: %s", outCombo)
	}

	// 3. Search non-matching term
	outNone, err := executeCommand("list", "-s", "nonexistentterm999")
	if err != nil {
		t.Fatalf("list -s nonexistent failed: %v", err)
	}
	if !strings.Contains(outNone, "No templates found matching") {
		t.Errorf("Expected empty search message, got: %s", outNone)
	}

	// 4. Search with JSON
	outJSON, err := executeCommand("list", "-s", "fiber", "--json")
	if err != nil {
		t.Fatalf("list -s fiber --json failed: %v", err)
	}
	var tpls []map[string]interface{}
	if err := json.Unmarshal([]byte(outJSON), &tpls); err != nil {
		t.Fatalf("Failed to parse search JSON: %v", err)
	}
	if len(tpls) == 0 {
		t.Errorf("Expected at least 1 template in search JSON, got 0")
	}
}

func TestAddCmd_AllFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal Go Fiber project
	goModContent := "module testalladdons\n\ngo 1.24\n\nrequire github.com/gofiber/fiber/v2 v2.52.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// 1. Run add --all
	out, err := executeCommand("add", "--all", "--dir", tempDir)
	if err != nil {
		t.Fatalf("add --all failed: %v", err)
	}
	if !strings.Contains(out, "Injected Addons:") {
		t.Errorf("Expected Injected Addons output, got: %s", out)
	}

	// Verify key files were generated
	if _, err := os.Stat(filepath.Join(tempDir, "Dockerfile")); os.IsNotExist(err) {
		t.Errorf("Expected Dockerfile to be created by add --all")
	}
	if _, err := os.Stat(filepath.Join(tempDir, ".github", "workflows", "ci.yml")); os.IsNotExist(err) {
		t.Errorf("Expected CI workflow to be created by add --all")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "internal", "database", "postgres.go")); os.IsNotExist(err) {
		t.Errorf("Expected postgres.go to be created by add --all")
	}

	// 2. Second run without force should report all installed
	out2, err := executeCommand("add", "--all", "--dir", tempDir)
	if err != nil {
		t.Fatalf("second add --all failed: %v", err)
	}
	if !strings.Contains(out2, "All available infrastructure addons are already installed") {
		t.Errorf("Expected all installed message on second run, got: %s", out2)
	}
}


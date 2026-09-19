package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"umaru/internal/config"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "umaru-test-config-*")
	if err == nil {
		config.SetTestConfigDir(tmpDir)
		defer os.RemoveAll(tmpDir)
	}
	os.Exit(m.Run())
}

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
	addDryRunFlag = false
	addSkipInstallFlag = false
	versionJSONFlag = false
	configJSONFlag = false
	listCategoryFlag = ""
	listSearchFlag = ""
	listJSONFlag = false
	quietFlag = false
	noColorFlag = false
	debugFlag = false
	configFileFlag = ""
	config.SetCustomConfigFile("")
	cleanDirFlag = "."
	cleanRecursiveFlag = false
	cleanDryRunFlag = false
	cleanForceFlag = false
	cleanAllFlag = false
	cleanJSONFlag = false
	devDirFlag = "."
	devPortFlag = ""
	devHostFlag = ""
	devPkgManagerFlag = ""
	devDryRunFlag = false
	devJSONFlag = false
	infoJSONFlag = false
	doctorJSONFlag = false

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

func TestVersionCmd_JSON(t *testing.T) {
	out, err := executeCommand("version", "--json")
	if err != nil {
		t.Fatalf("version --json failed: %v", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("version --json output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if info["version"] == nil || info["os"] == nil || info["arch"] == nil {
		t.Errorf("Expected version JSON to contain version, os, and arch fields: %v", info)
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

func TestListCmd_Alias(t *testing.T) {
	out, err := executeCommand("ls")
	if err != nil {
		t.Fatalf("ls command alias failed: %v", err)
	}

	if !strings.Contains(out, "Available Starter Templates") {
		t.Errorf("Expected ls output to contain 'Available Starter Templates', got: %s", out)
	}
	if !strings.Contains(out, "go-fiber") {
		t.Errorf("Expected ls output to contain 'go-fiber', got: %s", out)
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
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)

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
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)

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

func TestConfigCmd_Unset(t *testing.T) {
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)

	// 1. Set author
	_, err := executeCommand("config", "set", "author", "Original Author")
	if err != nil {
		t.Fatalf("config set author failed: %v", err)
	}

	// Verify author is set
	getOut, err := executeCommand("config", "get", "author")
	if err != nil {
		t.Fatalf("config get author failed: %v", err)
	}
	if !strings.Contains(getOut, "Original Author") {
		t.Fatalf("Expected 'Original Author', got: %s", getOut)
	}

	// 2. Unset author
	unsetOut, err := executeCommand("config", "unset", "author")
	if err != nil {
		t.Fatalf("config unset author failed: %v", err)
	}
	if !strings.Contains(unsetOut, "unset successfully") {
		t.Errorf("Expected unset confirmation message, got: %s", unsetOut)
	}

	// 3. Verify author is empty (default)
	getAfter, err := executeCommand("config", "get", "author")
	if err != nil {
		t.Fatalf("config get author after unset failed: %v", err)
	}
	if strings.TrimSpace(getAfter) != "" {
		t.Errorf("Expected empty author after unset, got: '%s'", getAfter)
	}

	// 4. Unset invalid key should return error
	_, err = executeCommand("config", "unset", "invalid-key-xyz")
	if err == nil {
		t.Errorf("Expected error unsetting invalid key, got nil")
	}
}

func TestConfigCmd_InvalidGitInit(t *testing.T) {
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)

	_, err := executeCommand("config", "set", "git-init", "not-a-boolean")
	if err == nil {
		t.Fatalf("Expected error when setting git-init to invalid value, got nil")
	}
	if !strings.Contains(err.Error(), "invalid boolean value") {
		t.Errorf("Expected error to mention 'invalid boolean value', got: %v", err)
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
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)

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
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)

	// 1. Set key
	setOut, err := executeCommand("config", "set", "author", "Test Engineer")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(setOut), "updated") {
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
	if !strings.Contains(out, "Injected") {
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

func TestAddCmd_DryRunAndSkipInstall(t *testing.T) {
	tempDir := t.TempDir()

	// Minimal Go project
	goModContent := "module testdryrunaddon\n\ngo 1.24\n\nrequire github.com/gofiber/fiber/v2 v2.52.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// 1. Dry run should NOT create redis.go
	outDryRun, err := executeCommand("add", "redis", "--dry-run", "--dir", tempDir)
	if err != nil {
		t.Fatalf("add redis --dry-run failed: %v", err)
	}
	if !strings.Contains(outDryRun, "Dry-Run Mode") {
		t.Errorf("Expected dry run header in output, got: %s", outDryRun)
	}
	redisPath := filepath.Join(tempDir, "internal", "cache", "redis.go")
	if _, err := os.Stat(redisPath); !os.IsNotExist(err) {
		t.Errorf("Expected %s to NOT exist after dry-run", redisPath)
	}

	// 2. Skip install should create the file without executing go get
	outSkip, err := executeCommand("add", "redis", "--skip-install", "--dir", tempDir)
	if err != nil {
		t.Fatalf("add redis --skip-install failed: %v", err)
	}
	if !strings.Contains(outSkip, "Addon(s) Injected Successfully") {
		t.Errorf("Expected success output, got: %s", outSkip)
	}
	if _, err := os.Stat(redisPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist after add with --skip-install", redisPath)
	}
}

func TestGlobalFlags_DebugAndConfig(t *testing.T) {
	tempDir := t.TempDir()
	customConfigPath := filepath.Join(tempDir, "custom-cfg.json")
	customConfig := `{"packageManager":"yarn","author":"Custom Tester","license":"Apache-2.0","gitInit":false}`
	if err := os.WriteFile(customConfigPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("Failed to write custom config: %v", err)
	}

	// 1. Test --config flag
	out, err := executeCommand("config", "get", "author", "--config", customConfigPath)
	if err != nil {
		t.Fatalf("config get author --config failed: %v", err)
	}
	if !strings.Contains(out, "Custom Tester") {
		t.Errorf("Expected custom author 'Custom Tester', got: %s", out)
	}

	// 2. Test --debug flag
	_, err = executeCommand("version", "--debug")
	if err != nil {
		t.Fatalf("version --debug failed: %v", err)
	}
	if !debugFlag {
		t.Errorf("Expected debugFlag to be true")
	}
	if !IsDebug() {
		t.Errorf("Expected IsDebug() to return true")
	}
}

func TestConfigInitCmd_NonInteractive(t *testing.T) {
	_, err := executeCommand("config", "init")
	if err == nil {
		t.Errorf("Expected config init to fail in non-terminal environment, but got nil")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Errorf("Expected error message to mention terminal, got: %v", err)
	}
}

func TestCleanCmd_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	nodeModulesDir := filepath.Join(tempDir, "node_modules", "testpkg")
	_ = os.MkdirAll(nodeModulesDir, 0755)
	_ = os.WriteFile(filepath.Join(nodeModulesDir, "index.js"), []byte("console.log('clean test')"), 0644)

	distDir := filepath.Join(tempDir, "dist")
	_ = os.MkdirAll(distDir, 0755)
	_ = os.WriteFile(filepath.Join(distDir, "output.js"), []byte("bundled code"), 0644)

	out, err := executeCommand("clean", "--dry-run", tempDir)
	if err != nil {
		t.Fatalf("clean --dry-run failed: %v", err)
	}

	if !strings.Contains(out, "node_modules") || !strings.Contains(out, "dist") {
		t.Errorf("Expected dry run output to list node_modules and dist, got:\n%s", out)
	}
	if !strings.Contains(out, "Dry-run mode") {
		t.Errorf("Expected dry run disclaimer, got:\n%s", out)
	}

	// Verify files still exist after dry-run
	if _, err := os.Stat(nodeModulesDir); os.IsNotExist(err) {
		t.Errorf("node_modules was deleted during dry run!")
	}
}

func TestCleanCmd_Force(t *testing.T) {
	tempDir := t.TempDir()
	nodeModulesDir := filepath.Join(tempDir, "node_modules", "testpkg")
	_ = os.MkdirAll(nodeModulesDir, 0755)
	_ = os.WriteFile(filepath.Join(nodeModulesDir, "index.js"), []byte("some code"), 0644)

	srcFile := filepath.Join(tempDir, "main.go")
	_ = os.WriteFile(srcFile, []byte("package main"), 0644)

	out, err := executeCommand("clean", "-f", tempDir)
	if err != nil {
		t.Fatalf("clean -f failed: %v", err)
	}

	if !strings.Contains(out, "Successfully reclaimed") {
		t.Errorf("Expected success output, got:\n%s", out)
	}

	// node_modules must be deleted
	if _, err := os.Stat(nodeModulesDir); !os.IsNotExist(err) {
		t.Errorf("node_modules still exists after clean -f")
	}

	// source code must remain intact
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Errorf("CRITICAL: main.go was deleted by clean -f!")
	}
}

func TestCleanCmd_JSON(t *testing.T) {
	tempDir := t.TempDir()
	distDir := filepath.Join(tempDir, "dist")
	_ = os.MkdirAll(distDir, 0755)
	_ = os.WriteFile(filepath.Join(distDir, "app.js"), []byte("compiled"), 0644)

	out, err := executeCommand("clean", "--dry-run", "--json", tempDir)
	if err != nil {
		t.Fatalf("clean --dry-run --json failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("clean --json output is not valid JSON: %v\nOutput: %s", err, out)
	}

	if data["items"] == nil || data["total_size_bytes"] == nil {
		t.Errorf("Expected JSON to have items and total_size_bytes fields: %v", data)
	}
}

func TestCleanCmd_AlreadyClean(t *testing.T) {
	tempDir := t.TempDir()
	out, err := executeCommand("clean", tempDir)
	if err != nil {
		t.Fatalf("clean on clean project failed: %v", err)
	}

	if !strings.Contains(out, "completely clean") {
		t.Errorf("Expected clean confirmation message, got: %s", out)
	}
}

func TestCleanCmd_Aliases(t *testing.T) {
	tempDir := t.TempDir()
	outSanitize, err := executeCommand("sanitize", tempDir)
	if err != nil {
		t.Fatalf("sanitize alias failed: %v", err)
	}
	if !strings.Contains(outSanitize, "Project Cleaner") {
		t.Errorf("Expected Project Cleaner header with sanitize, got: %s", outSanitize)
	}

	outPurge, err := executeCommand("purge", tempDir)
	if err != nil {
		t.Fatalf("purge alias failed: %v", err)
	}
	if !strings.Contains(outPurge, "Project Cleaner") {
		t.Errorf("Expected Project Cleaner header with purge, got: %s", outPurge)
	}
}

func TestDevCmd_DryRun_Go(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testgo\ngo 1.24\nrequire github.com/gofiber/fiber/v2 v2.52.0"), 0644)
	apiDir := filepath.Join(tempDir, "cmd", "api")
	_ = os.MkdirAll(apiDir, 0755)
	_ = os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main"), 0644)

	out, err := executeCommand("dev", "--dry-run", tempDir)
	if err != nil {
		t.Fatalf("dev --dry-run failed: %v", err)
	}

	if !strings.Contains(out, "Umaru Dev Runner") {
		t.Errorf("Expected Umaru Dev Runner header, got: %s", out)
	}
	if !strings.Contains(out, "go run cmd/api/main.go") {
		t.Errorf("Expected go run cmd/api/main.go command, got: %s", out)
	}
}

func TestDevCmd_DryRun_Node(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name":"myapp","scripts":{"dev":"vite"}}`), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "pnpm-lock.yaml"), []byte("lockfileVersion: 9.0"), 0644)

	out, err := executeCommand("dev", "--dry-run", "-p", "3000", tempDir)
	if err != nil {
		t.Fatalf("dev --dry-run failed: %v", err)
	}

	if !strings.Contains(out, "pnpm dev") {
		t.Errorf("Expected pnpm dev command, got: %s", out)
	}
	if !strings.Contains(out, "PORT=3000") {
		t.Errorf("Expected PORT=3000 in environment, got: %s", out)
	}
}

func TestDevCmd_DryRun_JSON(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "Cargo.toml"), []byte("[package]\nname = \"myrust\"\nversion = \"0.1.0\"\n[dependencies]\naxum = \"0.7\""), 0644)
	srcDir := filepath.Join(tempDir, "src")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte("fn main(){}"), 0644)

	out, err := executeCommand("dev", "--dry-run", "--json", tempDir)
	if err != nil {
		t.Fatalf("dev --dry-run --json failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("dev --json output is not valid JSON: %v\nOutput: %s", err, out)
	}

	if data["language"] != "rust" {
		t.Errorf("Expected language 'rust', got: %v", data["language"])
	}
	if data["command_line"] != "cargo run" {
		t.Errorf("Expected command_line 'cargo run', got: %v", data["command_line"])
	}
}

func TestDevCmd_Aliases(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testalias\ngo 1.24\nrequire github.com/gofiber/fiber/v2 v2.52.0"), 0644)
	apiDir := filepath.Join(tempDir, "cmd", "api")
	_ = os.MkdirAll(apiDir, 0755)
	_ = os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main"), 0644)

	outRun, err := executeCommand("run", "--dry-run", tempDir)
	if err != nil {
		t.Fatalf("run alias failed: %v", err)
	}
	if !strings.Contains(outRun, "go run cmd/api/main.go") {
		t.Errorf("Expected go run command with 'run' alias, got: %s", outRun)
	}

	outStart, err := executeCommand("start", "--dry-run", tempDir)
	if err != nil {
		t.Fatalf("start alias failed: %v", err)
	}
	if !strings.Contains(outStart, "go run cmd/api/main.go") {
		t.Errorf("Expected go run command with 'start' alias, got: %s", outStart)
	}
}

func TestConfigCmd_UnsetComprehensive(t *testing.T) {
	tempDir := t.TempDir()
	config.SetTestConfigDir(tempDir)
	defer config.SetTestConfigDir("")

	// 1. Unset author
	_, _ = executeCommand("config", "set", "author", "Test Author")
	out, err := executeCommand("config", "unset", "author")
	if err != nil {
		t.Fatalf("config unset author failed: %v", err)
	}
	if !strings.Contains(out, "Unset") && !strings.Contains(out, "author") {
		t.Errorf("Expected success output for unset author, got: %s", out)
	}

	// 2. Unset package-manager
	_, _ = executeCommand("config", "set", "package-manager", "pnpm")
	out, err = executeCommand("config", "unset", "package-manager")
	if err != nil {
		t.Fatalf("config unset package-manager failed: %v", err)
	}
	if !strings.Contains(out, "Unset") && !strings.Contains(out, "package-manager") {
		t.Errorf("Expected success output for unset package-manager, got: %s", out)
	}

	// 3. Unset license
	_, _ = executeCommand("config", "set", "license", "Apache-2.0")
	out, err = executeCommand("config", "unset", "license")
	if err != nil {
		t.Fatalf("config unset license failed: %v", err)
	}

	// 4. Unset git-init
	_, _ = executeCommand("config", "set", "git-init", "false")
	out, err = executeCommand("config", "unset", "git-init")
	if err != nil {
		t.Fatalf("config unset git-init failed: %v", err)
	}

	// 5. Unset invalid key
	_, err = executeCommand("config", "unset", "invalid-fake-key")
	if err == nil {
		t.Errorf("Expected error for unsetting invalid key, got nil")
	}

	// 6. Unset without arguments
	_, err = executeCommand("config", "unset")
	if err == nil {
		t.Errorf("Expected error for unset without args, got nil")
	}
}

func TestCleanCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Clean on already clean directory
	outClean, err := executeCommand("clean", "--dry-run", "--dir", tempDir)
	if err != nil {
		t.Fatalf("clean on clean directory failed: %v", err)
	}
	if !strings.Contains(outClean, "clean") && !strings.Contains(outClean, "Nothing") {
		t.Errorf("Expected clean directory message, got: %s", outClean)
	}

	// 2. Populate with removable targets
	nodeDir := filepath.Join(tempDir, "node_modules", "pkg")
	_ = os.MkdirAll(nodeDir, 0755)
	_ = os.WriteFile(filepath.Join(nodeDir, "index.js"), []byte("mod"), 0644)

	distDir := filepath.Join(tempDir, "dist")
	_ = os.MkdirAll(distDir, 0755)
	_ = os.WriteFile(filepath.Join(distDir, "app.js"), []byte("app"), 0644)

	// 3. Dry run test
	outDry, err := executeCommand("clean", "--dry-run", "--dir", tempDir)
	if err != nil {
		t.Fatalf("clean --dry-run failed: %v", err)
	}
	if !strings.Contains(outDry, "DRY-RUN") && !strings.Contains(outDry, "node_modules") {
		t.Errorf("Expected DRY-RUN notice in clean output, got: %s", outDry)
	}

	// 4. JSON dry run test
	outJSON, err := executeCommand("clean", "--dry-run", "--json", "--dir", tempDir)
	if err != nil {
		t.Fatalf("clean --json failed: %v", err)
	}
	var cleanData map[string]interface{}
	if err := json.Unmarshal([]byte(outJSON), &cleanData); err != nil {
		t.Fatalf("clean --json is not valid JSON: %v\nOutput: %s", err, outJSON)
	}
	if cleanData["root_dir"] == nil || cleanData["items"] == nil {
		t.Errorf("Expected root_dir and items in clean JSON output, got: %v", cleanData)
	}

	// 5. Force clean test
	outForce, err := executeCommand("clean", "-f", "--dir", tempDir)
	if err != nil {
		t.Fatalf("clean -f failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(outForce), "reclaimed") && !strings.Contains(strings.ToLower(outForce), "cleaned") {
		t.Errorf("Expected reclaimed message, got: %s", outForce)
	}

	// Verify folders are deleted
	if _, err := os.Stat(distDir); !os.IsNotExist(err) {
		t.Errorf("Expected dist to be deleted by clean -f")
	}
}

func TestDevCmd_Errors(t *testing.T) {
	// 1. Non-existent directory
	_, err := executeCommand("dev", "--dir", "non-existent-dev-dir-xyz")
	if err == nil {
		t.Errorf("Expected error for non-existent dir, got nil")
	}

	// 2. Unknown directory structure
	tempEmpty := t.TempDir()
	_, err = executeCommand("dev", "--dir", tempEmpty)
	if err == nil {
		t.Errorf("Expected error for empty dir with no supported stack, got nil")
	}
}

func TestCompletionCmd_Shells(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, sh := range shells {
		out, err := executeCommand("completion", sh)
		if err != nil {
			t.Errorf("completion %s failed: %v", sh, err)
		}
		if len(out) == 0 {
			t.Errorf("completion %s returned empty output", sh)
		}
	}

	// Invalid shell
	_, err := executeCommand("completion", "invalidshell")
	if err == nil {
		t.Errorf("Expected error for invalid shell, got nil")
	}
}

func TestInfoCmd_JSON_AndErrors(t *testing.T) {
	// 1. Valid info with --json
	out, err := executeCommand("info", "go-fiber", "--json")
	if err != nil {
		t.Fatalf("info go-fiber --json failed: %v", err)
	}

	var infoData map[string]interface{}
	if err := json.Unmarshal([]byte(out), &infoData); err != nil {
		t.Fatalf("info --json output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if infoData["Config"] == nil {
		t.Errorf("Expected 'Config' in info --json output, got: %v", infoData)
	}

	// 2. Non-existent template error
	_, err = executeCommand("info", "non-existent-fake-template-xyz")
	if err == nil {
		t.Errorf("Expected error for non-existent template in info, got nil")
	}
}

func TestDetectShell(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer func() {
		_ = os.Setenv("SHELL", origShell)
	}()

	_ = os.Setenv("SHELL", "/bin/zsh")
	if sh := detectShell(); sh != "zsh" {
		t.Errorf("Expected zsh, got %s", sh)
	}

	_ = os.Setenv("SHELL", "/usr/bin/bash")
	if sh := detectShell(); sh != "bash" {
		t.Errorf("Expected bash, got %s", sh)
	}

	_ = os.Setenv("SHELL", "/usr/local/bin/fish")
	if sh := detectShell(); sh != "fish" {
		t.Errorf("Expected fish, got %s", sh)
	}
}

func TestGetPowerShellProfilePath(t *testing.T) {
	dummyHome := filepath.Join(t.TempDir(), "fakehome")
	completionTestHomeDir = dummyHome
	defer func() {
		completionTestHomeDir = ""
	}()

	path := getPowerShellProfilePath(dummyHome)
	expected := filepath.Join(dummyHome, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestInitCmd_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "dryrun-project")

	out, err := executeCommand("init", targetPath, "-t", "go-fiber", "--no-git", "--skip-install", "--no-addons", "--dry-run")
	if err != nil {
		t.Fatalf("init --dry-run failed: %v", err)
	}
	if !strings.Contains(out, "Dry-Run") && !strings.Contains(out, "Simulation") {
		t.Errorf("Expected dry run output for init, got: %s", out)
	}
}

func TestInitCmd_ScaffoldWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "scaffold-project")

	out, err := executeCommand("init", targetPath, "-t", "go-fiber", "--no-git", "--skip-install", "--no-addons", "-f")
	if err != nil {
		t.Fatalf("init scaffold workflow failed: %v", err)
	}

	if !strings.Contains(out, "Project Scaffolding Complete") && !strings.Contains(strings.ToLower(out), "next steps") {
		t.Errorf("Expected success summary card in init output, got: %s", out)
	}

	// Verify project files were generated
	if _, err := os.Stat(filepath.Join(targetPath, "go.mod")); os.IsNotExist(err) {
		t.Errorf("Expected go.mod to be created in %s", targetPath)
	}
	if _, err := os.Stat(filepath.Join(targetPath, "cmd", "api", "main.go")); os.IsNotExist(err) {
		t.Errorf("Expected cmd/api/main.go to be created in %s", targetPath)
	}
}








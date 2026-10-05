package testrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"umaru/internal/config"
	"umaru/internal/generator"
)

// TestOptions specifies user options for discovering and executing tests
type TestOptions struct {
	TargetDir      string
	Verbose        bool   // verbose test output (-v)
	Watch          bool   // watch mode / continuous testing
	Coverage       bool   // generate code coverage
	Filter         string // run only tests matching regex/substring
	Race           bool   // race detection (Go)
	PackageManager string // optional override (npm, pnpm, yarn, bun)
}

// TestConfig describes the resolved test command and environment
type TestConfig struct {
	TargetDir   string   `json:"target_dir"`
	Language    string   `json:"language"`
	Framework   string   `json:"framework"`
	Command     []string `json:"command"`
	CommandLine string   `json:"command_line"`
	Env         []string `json:"env,omitempty"`
}

// fileExists checks if a path exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// findPythonVenvExecutable looks for an executable inside a local virtual environment (.venv / venv)
func findPythonVenvExecutable(baseDir, exeName string) string {
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			filepath.Join(baseDir, ".venv", "Scripts", exeName+".exe"),
			filepath.Join(baseDir, "venv", "Scripts", exeName+".exe"),
		}
	} else {
		candidates = []string{
			filepath.Join(baseDir, ".venv", "bin", exeName),
			filepath.Join(baseDir, "venv", "bin", exeName),
		}
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

// detectNodePackageManager determines the package manager based on lockfiles or user config
func detectNodePackageManager(dir string, override string) string {
	if override != "" {
		return strings.ToLower(strings.TrimSpace(override))
	}

	if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock")) {
		return "bun"
	}
	if fileExists(filepath.Join(dir, "yarn.lock")) {
		return "yarn"
	}
	if fileExists(filepath.Join(dir, "package-lock.json")) {
		return "npm"
	}

	userCfg := config.LoadUserConfig()
	if userCfg.PackageManager != "" {
		return userCfg.PackageManager
	}

	return "npm"
}

// DetectTestCommand inspects the target directory and resolves the optimal test command
func DetectTestCommand(opts TestOptions) (*TestConfig, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}

	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("directory error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target path is not a directory: %s", absDir)
	}

	detected, err := generator.DetectProject(absDir)
	if err != nil {
		return nil, err
	}

	cfg := &TestConfig{
		TargetDir: absDir,
		Language:  string(detected.Type),
		Framework: detected.Framework,
	}

	// 1. Go Projects
	if detected.Type == generator.ProjectTypeGo {
		cmd := []string{"go", "test"}
		if opts.Verbose {
			cmd = append(cmd, "-v")
		}
		if opts.Race {
			cmd = append(cmd, "-race")
		}
		if opts.Coverage {
			cmd = append(cmd, "-coverprofile=coverage.out", "-cover")
		}
		if opts.Filter != "" {
			cmd = append(cmd, "-run", opts.Filter)
		}
		cmd = append(cmd, "./...")

		cfg.Command = cmd
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 2. Node.js / TypeScript Projects
	if detected.Type == generator.ProjectTypeNode {
		pm := detectNodePackageManager(absDir, opts.PackageManager)
		pkgPath := filepath.Join(absDir, "package.json")
		data, err := os.ReadFile(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("could not read package.json: %w", err)
		}

		var pkgMap struct {
			Scripts      map[string]string `json:"scripts"`
			Dependencies map[string]string `json:"dependencies"`
			DevDeps      map[string]string `json:"devDependencies"`
		}
		_ = json.Unmarshal(data, &pkgMap)

		hasTestScript := false
		if script, ok := pkgMap.Scripts["test"]; ok {
			trimmed := strings.TrimSpace(script)
			if trimmed != "" && !strings.Contains(trimmed, "no test specified") {
				hasTestScript = true
			}
		}

		hasWatchScript := false
		if _, ok := pkgMap.Scripts["test:watch"]; ok {
			hasWatchScript = true
		}

		hasCoverageScript := false
		if _, ok := pkgMap.Scripts["test:coverage"]; ok {
			hasCoverageScript = true
		}

		// Check for testing tools in deps
		hasVitest := pkgMap.DevDeps["vitest"] != "" || pkgMap.Dependencies["vitest"] != ""
		hasJest := pkgMap.DevDeps["jest"] != "" || pkgMap.Dependencies["jest"] != ""

		var cmd []string

		if opts.Watch && hasWatchScript {
			cmd = []string{pm, "run", "test:watch"}
		} else if opts.Coverage && hasCoverageScript {
			cmd = []string{pm, "run", "test:coverage"}
		} else if hasTestScript {
			cmd = []string{pm, "test"}
			var extraArgs []string
			if opts.Watch {
				extraArgs = append(extraArgs, "--watch")
			}
			if opts.Coverage {
				extraArgs = append(extraArgs, "--coverage")
			}
			if opts.Filter != "" {
				extraArgs = append(extraArgs, "-t", opts.Filter)
			}
			if len(extraArgs) > 0 {
				if pm == "npm" {
					cmd = append(cmd, "--")
				}
				cmd = append(cmd, extraArgs...)
			}
		} else if pm == "bun" {
			cmd = []string{"bun", "test"}
			if opts.Watch {
				cmd = append(cmd, "--watch")
			}
			if opts.Coverage {
				cmd = append(cmd, "--coverage")
			}
			if opts.Filter != "" {
				cmd = append(cmd, opts.Filter)
			}
		} else if hasVitest {
			cmd = []string{"npx", "vitest"}
			if !opts.Watch {
				cmd = append(cmd, "run")
			}
			if opts.Coverage {
				cmd = append(cmd, "--coverage")
			}
			if opts.Filter != "" {
				cmd = append(cmd, "-t", opts.Filter)
			}
		} else if hasJest {
			cmd = []string{"npx", "jest"}
			if opts.Watch {
				cmd = append(cmd, "--watch")
			}
			if opts.Coverage {
				cmd = append(cmd, "--coverage")
			}
			if opts.Filter != "" {
				cmd = append(cmd, "-t", opts.Filter)
			}
		} else {
			// Default fallback
			cmd = []string{pm, "test"}
		}

		cfg.Command = cmd
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 3. Python Projects
	if detected.Type == generator.ProjectTypePython {
		venvPytest := findPythonVenvExecutable(absDir, "pytest")
		venvPython := findPythonVenvExecutable(absDir, "python")

		pytestBin := "pytest"
		if venvPytest != "" {
			pytestBin = venvPytest
		}

		// Check if pytest is available or fallback to unittest
		usePytest := false
		if venvPytest != "" {
			usePytest = true
		} else if _, err := exec.LookPath("pytest"); err == nil {
			usePytest = true
		}

		if usePytest {
			cmd := []string{pytestBin}
			if opts.Verbose {
				cmd = append(cmd, "-v")
			}
			if opts.Coverage {
				cmd = append(cmd, "--cov=.")
			}
			if opts.Filter != "" {
				cmd = append(cmd, "-k", opts.Filter)
			}
			cfg.Command = cmd
		} else {
			pyBin := "python"
			if venvPython != "" {
				pyBin = venvPython
			}
			cmd := []string{pyBin, "-m", "unittest"}
			if opts.Verbose {
				cmd = append(cmd, "-v")
			}
			if opts.Filter != "" {
				cmd = append(cmd, "-k", opts.Filter)
			}
			cfg.Command = cmd
		}

		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 4. Rust Projects
	if detected.Type == generator.ProjectTypeRust {
		cmd := []string{"cargo", "test"}
		if opts.Filter != "" {
			cmd = append(cmd, opts.Filter)
		}
		if opts.Verbose {
			cmd = append(cmd, "--", "--nocapture")
		}
		cfg.Command = cmd
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	return nil, errors.New("unable to determine test command for this project structure")
}

// Run executes the test command streaming standard input, output, and error
func Run(ctx context.Context, cfg *TestConfig) error {
	if len(cfg.Command) == 0 {
		return errors.New("no command specified")
	}

	cmd := exec.CommandContext(ctx, cfg.Command[0], cfg.Command[1:]...)
	cmd.Dir = cfg.TargetDir
	cmd.Env = append(os.Environ(), cfg.Env...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

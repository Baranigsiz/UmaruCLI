package devrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"umaru/internal/config"
	"umaru/internal/generator"
)

// DevOptions specifies configuration for detecting and executing the dev command
type DevOptions struct {
	TargetDir      string
	Port           string
	Host           string
	PackageManager string // optional override (npm, pnpm, yarn, bun)
}

// DevConfig contains all information necessary to execute the development server
type DevConfig struct {
	TargetDir   string   `json:"target_dir"`
	Language    string   `json:"language"`
	Framework   string   `json:"framework"`
	Command     []string `json:"command"`
	CommandLine string   `json:"command_line"`
	Env         []string `json:"env"`
	Port        string   `json:"port,omitempty"`
	Host        string   `json:"host,omitempty"`
}

// fileExists checks if a path exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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

// DetectDevCommand inspects the target directory and resolves the optimal development command
func DetectDevCommand(opts DevOptions) (*DevConfig, error) {
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

	cfg := &DevConfig{
		TargetDir: absDir,
		Language:  string(detected.Type),
		Framework: detected.Framework,
		Port:      opts.Port,
		Host:      opts.Host,
	}

	var envList []string
	if opts.Port != "" {
		envList = append(envList, fmt.Sprintf("PORT=%s", opts.Port))
	}
	if opts.Host != "" {
		envList = append(envList, fmt.Sprintf("HOST=%s", opts.Host))
	}
	cfg.Env = envList

	// 0. Monorepo (Fullstack Go + React)
	if detected.Framework == "fullstack-go-react" {
		cfg.Command = []string{"docker-compose", "up", "--build"}
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 1. Go Projects
	if detected.Type == generator.ProjectTypeGo {
		if fileExists(filepath.Join(absDir, "cmd", "api", "main.go")) {
			cfg.Command = []string{"go", "run", "cmd/api/main.go"}
		} else if fileExists(filepath.Join(absDir, "cmd", "web", "main.go")) {
			cfg.Command = []string{"go", "run", "cmd/web/main.go"}
		} else if fileExists(filepath.Join(absDir, "main.go")) {
			cfg.Command = []string{"go", "run", "main.go"}
		} else {
			cfg.Command = []string{"go", "run", "."}
		}
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 2. Node.js / JavaScript / TypeScript Projects
	if detected.Type == generator.ProjectTypeNode {
		pm := detectNodePackageManager(absDir, opts.PackageManager)
		pkgPath := filepath.Join(absDir, "package.json")
		data, err := os.ReadFile(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("could not read package.json: %w", err)
		}

		var pkgMap struct {
			Scripts map[string]string `json:"scripts"`
		}
		_ = json.Unmarshal(data, &pkgMap)

		scriptName := "dev"
		if _, ok := pkgMap.Scripts["desktop:dev"]; ok && detected.Framework == "tauri-desktop" {
			scriptName = "desktop:dev"
		} else if _, ok := pkgMap.Scripts["start:dev"]; ok {
			scriptName = "start:dev"
		} else if _, ok := pkgMap.Scripts["dev"]; ok {
			scriptName = "dev"
		} else if _, ok := pkgMap.Scripts["start"]; ok {
			scriptName = "start"
		}

		switch pm {
		case "pnpm":
			cfg.Command = []string{"pnpm", scriptName}
		case "yarn":
			cfg.Command = []string{"yarn", scriptName}
		case "bun":
			if scriptName == "start" {
				cfg.Command = []string{"bun", "start"}
			} else {
				cfg.Command = []string{"bun", "run", scriptName}
			}
		default: // npm
			if scriptName == "start" {
				cfg.Command = []string{"npm", "start"}
			} else {
				cfg.Command = []string{"npm", "run", scriptName}
			}
		}

		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 3. Python Projects
	if detected.Type == generator.ProjectTypePython {
		if fileExists(filepath.Join(absDir, "app", "main.py")) {
			cmd := []string{"uvicorn", "app.main:app", "--reload"}
			if opts.Port != "" {
				cmd = append(cmd, "--port", opts.Port)
			}
			if opts.Host != "" {
				cmd = append(cmd, "--host", opts.Host)
			}
			cfg.Command = cmd
		} else if fileExists(filepath.Join(absDir, "main.py")) {
			cfg.Command = []string{"python", "main.py"}
		} else {
			cfg.Command = []string{"python", "-m", "app.main"}
		}
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	// 4. Rust Projects
	if detected.Type == generator.ProjectTypeRust {
		cfg.Command = []string{"cargo", "run"}
		cfg.CommandLine = strings.Join(cfg.Command, " ")
		return cfg, nil
	}

	return nil, errors.New("unable to determine dev command for this project structure")
}

// Run executes the development command with live interactive stdio
func Run(ctx context.Context, cfg *DevConfig) error {
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

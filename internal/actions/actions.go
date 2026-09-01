package actions

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// buildCommand creates a cross-platform exec.Cmd
func buildCommand(dir string, command []string) *exec.Cmd {
	if len(command) == 0 {
		return nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, package managers like npm/pnpm/yarn/bun are cmd/bat scripts.
		// Using cmd.exe /c ensures proper PATH resolution and execution.
		fullCmd := strings.Join(command, " ")
		cmd = exec.Command("cmd.exe", "/c", fullCmd)
	} else {
		cmd = exec.Command(command[0], command[1:]...)
	}

	cmd.Dir = dir
	return cmd
}

// InitGit initializes a git repository in the given directory
func InitGit(projectPath string) error {
	cmd := buildCommand(projectPath, []string{"git", "init"})
	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git init failed: %s", outStr)
		}
		return fmt.Errorf("git init failed: %w", err)
	}
	return nil
}

// InstallDependencies runs the specified package manager installation command
func InstallDependencies(projectPath string, installCommand []string, verbose bool) error {
	if len(installCommand) == 0 {
		return nil
	}

	cmd := buildCommand(projectPath, installCommand)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s failed: %w", strings.Join(installCommand, " "), err)
		}
		return nil
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("%s failed:\n%s", strings.Join(installCommand, " "), outStr)
		}
		return fmt.Errorf("%s failed: %w", strings.Join(installCommand, " "), err)
	}
	return nil
}


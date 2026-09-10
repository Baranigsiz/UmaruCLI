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
		// On Windows, JS package managers (npm, pnpm, yarn, bun) are cmd/bat scripts and need cmd.exe /c.
		// Direct executables (git, go, etc.) should run directly so Go handles argument quoting natively.
		first := strings.ToLower(command[0])
		if first == "npm" || first == "pnpm" || first == "yarn" || first == "bun" {
			fullCmd := strings.Join(command, " ")
			cmd = exec.Command("cmd.exe", "/c", fullCmd)
		} else {
			cmd = exec.Command(command[0], command[1:]...)
		}
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

// CommitGit stages all files and creates an initial commit
func CommitGit(projectPath string, message string) error {
	addCmd := buildCommand(projectPath, []string{"git", "add", "-A"})
	if out, err := addCmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git add failed: %s", outStr)
		}
		return fmt.Errorf("git add failed: %w", err)
	}

	commitCmd := buildCommand(projectPath, []string{
		"git",
		"-c", "user.name=Umaru CLI",
		"-c", "user.email=umaru@cli.local",
		"commit",
		"-m", message,
	})
	if out, err := commitCmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git commit failed: %s", outStr)
		}
		return fmt.Errorf("git commit failed: %w", err)
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


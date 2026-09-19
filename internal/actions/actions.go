package actions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// buildCommandContext creates a cross-platform exec.Cmd with context support
func buildCommandContext(ctx context.Context, dir string, command []string) *exec.Cmd {
	if len(command) == 0 {
		return nil
	}

	execCmd := make([]string, len(command))
	copy(execCmd, command)

	// Fallback pip to pip3 if pip is not found in PATH
	if strings.ToLower(execCmd[0]) == "pip" {
		if _, err := exec.LookPath("pip"); err != nil {
			if _, err3 := exec.LookPath("pip3"); err3 == nil {
				execCmd[0] = "pip3"
			}
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, JS package managers (npm, pnpm, yarn, bun) are cmd/bat scripts and need cmd.exe /c.
		// Direct executables (git, go, etc.) should run directly so Go handles argument quoting natively.
		first := strings.ToLower(execCmd[0])
		if first == "npm" || first == "pnpm" || first == "yarn" || first == "bun" {
			var quotedArgs []string
			for _, arg := range execCmd {
				if strings.Contains(arg, " ") && !strings.HasPrefix(arg, "\"") {
					quotedArgs = append(quotedArgs, fmt.Sprintf("%q", arg))
				} else {
					quotedArgs = append(quotedArgs, arg)
				}
			}
			cmd = exec.CommandContext(ctx, "cmd.exe", "/c", strings.Join(quotedArgs, " "))
		} else {
			cmd = exec.CommandContext(ctx, execCmd[0], execCmd[1:]...)
		}
	} else {
		cmd = exec.CommandContext(ctx, execCmd[0], execCmd[1:]...)
	}

	cmd.Dir = dir
	return cmd
}

// buildCommand creates a cross-platform exec.Cmd
func buildCommand(dir string, command []string) *exec.Cmd {
	return buildCommandContext(context.Background(), dir, command)
}

// InitGitContext initializes a git repository in the given directory with context
func InitGitContext(ctx context.Context, projectPath string) error {
	cmd := buildCommandContext(ctx, projectPath, []string{"git", "init"})
	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git init failed: %s", outStr)
		}
		return fmt.Errorf("git init failed: %w", err)
	}
	return nil
}

// InitGit initializes a git repository in the given directory
func InitGit(projectPath string) error {
	return InitGitContext(context.Background(), projectPath)
}

// CommitGitContext stages all files and creates an initial commit with context
func CommitGitContext(ctx context.Context, projectPath string, message string, author ...string) error {
	addCmd := buildCommandContext(ctx, projectPath, []string{"git", "add", "-A"})
	if out, err := addCmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git add failed: %s", outStr)
		}
		return fmt.Errorf("git add failed: %w", err)
	}

	// Check if user has global or local git identity configured
	hasGitUser := false
	checkName := exec.CommandContext(ctx, "git", "config", "user.name")
	if nameOut, err := checkName.Output(); err == nil && len(strings.TrimSpace(string(nameOut))) > 0 {
		checkEmail := exec.CommandContext(ctx, "git", "config", "user.email")
		if emailOut, err := checkEmail.Output(); err == nil && len(strings.TrimSpace(string(emailOut))) > 0 {
			hasGitUser = true
		}
	}

	var commitArgs []string
	if hasGitUser {
		// Use developer's native Git author identity and signing settings
		commitArgs = []string{"git", "commit", "-m", message}
	} else {
		var authorName string
		if len(author) > 0 {
			authorName = strings.TrimSpace(author[0])
		}
		if authorName != "" {
			commitArgs = []string{
				"git",
				"-c", fmt.Sprintf("user.name=%s", authorName),
				"-c", fmt.Sprintf("user.email=%s@users.noreply.local", strings.ToLower(strings.ReplaceAll(authorName, " ", "-"))),
				"commit",
				"-m", message,
			}
		} else {
			commitArgs = []string{
				"git",
				"-c", "user.name=Umaru CLI",
				"-c", "user.email=umaru@cli.local",
				"commit",
				"-m", message,
			}
		}
	}

	commitCmd := buildCommandContext(ctx, projectPath, commitArgs)
	if out, err := commitCmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git commit failed: %s", outStr)
		}
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

// CommitGit stages all files and creates an initial commit
func CommitGit(projectPath string, message string, author ...string) error {
	return CommitGitContext(context.Background(), projectPath, message, author...)
}

// InstallDependenciesContext runs the specified package manager installation command with context
func InstallDependenciesContext(ctx context.Context, projectPath string, installCommand []string, verbose bool) error {
	if len(installCommand) == 0 {
		return nil
	}

	cmd := buildCommandContext(ctx, projectPath, installCommand)

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

// InstallDependencies runs the specified package manager installation command
func InstallDependencies(projectPath string, installCommand []string, verbose bool) error {
	return InstallDependenciesContext(context.Background(), projectPath, installCommand, verbose)
}


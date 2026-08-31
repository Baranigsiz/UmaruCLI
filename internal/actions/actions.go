package actions

import (
	"fmt"
	"os/exec"
	"strings"
	"umaru/internal/templates"
)

// InitGit initializes a git repository in the given directory
func InitGit(projectPath string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("git init failed: %s", outStr)
		}
		return fmt.Errorf("git init failed: %w", err)
	}
	return nil
}

// InstallDependencies runs the appropriate package manager based on the template
func InstallDependencies(projectPath string, template templates.TemplateConfig) error {
	if len(template.InstallCommand) == 0 {
		return nil
	}

	cmd := exec.Command(template.InstallCommand[0], template.InstallCommand[1:]...)
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return fmt.Errorf("%s failed:\n%s", strings.Join(template.InstallCommand, " "), outStr)
		}
		return fmt.Errorf("%s failed: %w", strings.Join(template.InstallCommand, " "), err)
	}
	return nil
}

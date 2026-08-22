package actions

import (
	"os/exec"
	"umaru/internal/templates"
)

// InitGit initializes a git repository in the given directory
func InitGit(projectPath string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = projectPath
	return cmd.Run()
}

// InstallDependencies runs the appropriate package manager based on the template
func InstallDependencies(projectPath string, template templates.TemplateConfig) error {
	if len(template.InstallCommand) == 0 {
		return nil
	}

	cmd := exec.Command(template.InstallCommand[0], template.InstallCommand[1:]...)
	cmd.Dir = projectPath
	return cmd.Run()
}

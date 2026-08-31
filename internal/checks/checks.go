package checks

import (
	"fmt"
	"os/exec"
	"umaru/internal/templates"
)

// PreFlightChecks verifies that all required system tools are available based on execution options
func PreFlightChecks(template templates.TemplateConfig, checkGit bool, checkInstall bool) error {
	// Check for git if git initialization is requested
	if checkGit {
		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("'git' is not installed or not found in PATH")
		}
	}

	// Check for the package manager required by the template if installation is requested
	if checkInstall && len(template.InstallCommand) > 0 {
		pkgManager := template.InstallCommand[0]
		if _, err := exec.LookPath(pkgManager); err != nil {
			return fmt.Errorf("'%s' is required for this template but was not found in PATH", pkgManager)
		}
	}

	return nil
}

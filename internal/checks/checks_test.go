package checks

import (
	"os/exec"
	"testing"
	"umaru/internal/templates"
)

func TestPreFlightChecks(t *testing.T) {
	// First, check if 'git' is actually in the system PATH.
	// If it's not, even the success case will fail, so we should skip or mock.
	// For this unit test, we assume 'go' and 'git' exist on the machine running tests.
	_, gitErr := exec.LookPath("git")
	_, goErr := exec.LookPath("go")
	if gitErr != nil || goErr != nil {
		t.Skip("Skipping test because 'git' or 'go' is not available in PATH")
	}

	tests := []struct {
		name         string
		template     templates.TemplateConfig
		checkGit     bool
		checkInstall bool
		expectError  bool
	}{
		{
			name: "Valid Template (go)",
			template: templates.TemplateConfig{
				InstallCommand: []string{"go", "mod", "tidy"},
			},
			checkGit:     true,
			checkInstall: true,
			expectError:  false,
		},
		{
			name: "Invalid Template (non-existent command)",
			template: templates.TemplateConfig{
				InstallCommand: []string{"some_fake_command_123"},
			},
			checkGit:     true,
			checkInstall: true,
			expectError:  true,
		},
		{
			name: "Invalid Template but checkInstall is false",
			template: templates.TemplateConfig{
				InstallCommand: []string{"some_fake_command_123"},
			},
			checkGit:     true,
			checkInstall: false,
			expectError:  false,
		},
		{
			name: "Empty Install Command",
			template: templates.TemplateConfig{
				InstallCommand: []string{},
			},
			checkGit:     true,
			checkInstall: true,
			expectError:  false, // Should pass because there's no pkgManager to check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PreFlightChecks(tt.template, tt.checkGit, tt.checkInstall)
			if (err != nil) != tt.expectError {
				t.Errorf("PreFlightChecks() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

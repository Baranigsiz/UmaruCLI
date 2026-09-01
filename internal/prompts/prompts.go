package prompts

import (
	"fmt"
	"strings"
	"umaru/internal/templates"

	"github.com/charmbracelet/huh"
)

type PromptResult struct {
	ProjectName    string
	Template       templates.TemplateConfig
	PackageManager string
}

func Run(initialName string, initialTemplateID string, initialPkgManager string) (*PromptResult, error) {
	projectName := strings.TrimSpace(initialName)
	selectedTemplateID := strings.TrimSpace(initialTemplateID)
	selectedPkgManager := strings.TrimSpace(initialPkgManager)

	// Dynamically load templates
	availableTemplates, err := templates.GetAvailableTemplates()
	if err != nil || len(availableTemplates) == 0 {
		return nil, fmt.Errorf("could not load templates: %v", err)
	}

	// If template ID is already provided, validate it
	var preSelectedTmpl *templates.TemplateConfig
	if selectedTemplateID != "" {
		tmpl, err := templates.FindTemplateByID(selectedTemplateID)
		if err != nil {
			return nil, err
		}
		preSelectedTmpl = tmpl
	}

	// If project name and template ID are provided (and pkgManager if node-based or already provided), return immediately
	if projectName != "" && preSelectedTmpl != nil {
		if !preSelectedTmpl.IsNodeBased() || selectedPkgManager != "" {
			return &PromptResult{
				ProjectName:    projectName,
				Template:       *preSelectedTmpl,
				PackageManager: selectedPkgManager,
			}, nil
		}
	}

	// 1. First Group: Project Name & Template Selection
	var basicFields []huh.Field

	if projectName == "" {
		basicFields = append(basicFields,
			huh.NewInput().
				Title("What is your project named?").
				Value(&projectName).
				Validate(func(str string) error {
					if len(strings.TrimSpace(str)) == 0 {
						return fmt.Errorf("project name cannot be empty")
					}
					return nil
				}),
		)
	}

	if selectedTemplateID == "" {
		var options []huh.Option[string]
		for _, t := range availableTemplates {
			options = append(options, huh.NewOption(fmt.Sprintf("%s - %s", t.Name, t.Description), t.ID))
		}

		basicFields = append(basicFields,
			huh.NewSelect[string]().
				Title("Choose a project template").
				Options(options...).
				Value(&selectedTemplateID),
		)
	}

	if len(basicFields) > 0 {
		form := huh.NewForm(huh.NewGroup(basicFields...))
		if err := form.Run(); err != nil {
			return nil, err
		}
	}

	// Resolve the selected template config
	tmpl, err := templates.FindTemplateByID(selectedTemplateID)
	if err != nil {
		return nil, err
	}

	// 2. Second Group: If template is Node-based and package manager is not specified, ask for it
	if tmpl.IsNodeBased() && selectedPkgManager == "" {
		pkgOptions := []huh.Option[string]{
			huh.NewOption("npm (Standard Node Package Manager)", "npm"),
			huh.NewOption("pnpm (Fast, disk space efficient)", "pnpm"),
			huh.NewOption("yarn (Classic Yarn Package Manager)", "yarn"),
			huh.NewOption("bun (Ultra-fast all-in-one JavaScript runtime)", "bun"),
		}

		pkgSelect := huh.NewSelect[string]().
			Title("Choose a package manager").
			Options(pkgOptions...).
			Value(&selectedPkgManager)

		pkgForm := huh.NewForm(huh.NewGroup(pkgSelect))
		if err := pkgForm.Run(); err != nil {
			return nil, err
		}
	}

	return &PromptResult{
		ProjectName:    strings.TrimSpace(projectName),
		Template:       *tmpl,
		PackageManager: selectedPkgManager,
	}, nil
}

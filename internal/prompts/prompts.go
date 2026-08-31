package prompts

import (
	"fmt"
	"strings"
	"umaru/internal/templates"

	"github.com/charmbracelet/huh"
)

type PromptResult struct {
	ProjectName string
	Template    templates.TemplateConfig
}

func Run(initialName string, initialTemplateID string) (*PromptResult, error) {
	projectName := strings.TrimSpace(initialName)
	selectedTemplateID := strings.TrimSpace(initialTemplateID)

	// Dynamically load templates
	availableTemplates, err := templates.GetAvailableTemplates()
	if err != nil || len(availableTemplates) == 0 {
		return nil, fmt.Errorf("could not load templates: %v", err)
	}

	// If template ID is already provided, validate it
	if selectedTemplateID != "" {
		tmpl, err := templates.FindTemplateByID(selectedTemplateID)
		if err != nil {
			return nil, err
		}
		// If project name is also provided, we're done (non-interactive)
		if projectName != "" {
			return &PromptResult{
				ProjectName: projectName,
				Template:    *tmpl,
			}, nil
		}
	}

	// Build fields based on what is missing
	var fields []huh.Field

	if projectName == "" {
		fields = append(fields,
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

		fields = append(fields,
			huh.NewSelect[string]().
				Title("Choose a project template").
				Options(options...).
				Value(&selectedTemplateID),
		)
	}

	if len(fields) > 0 {
		form := huh.NewForm(huh.NewGroup(fields...))
		err = form.Run()
		if err != nil {
			return nil, err
		}
	}

	// Find the selected template config
	tmpl, err := templates.FindTemplateByID(selectedTemplateID)
	if err != nil {
		return nil, err
	}

	return &PromptResult{
		ProjectName: strings.TrimSpace(projectName),
		Template:    *tmpl,
	}, nil
}

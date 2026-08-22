package prompts

import (
	"fmt"
	"umaru/internal/templates"

	"github.com/charmbracelet/huh"
)

type PromptResult struct {
	ProjectName    string
	Template       templates.TemplateConfig
}

func Run() (*PromptResult, error) {
	var projectName string
	var selectedTemplateID string

	// Dynamically load templates
	availableTemplates, err := templates.GetAvailableTemplates()
	if err != nil || len(availableTemplates) == 0 {
		return nil, fmt.Errorf("could not load templates: %v", err)
	}

	var options []huh.Option[string]
	for _, t := range availableTemplates {
		options = append(options, huh.NewOption(t.Name, t.ID))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What is your project named?").
				Value(&projectName).
				Validate(func(str string) error {
					if len(str) == 0 {
						return fmt.Errorf("project name cannot be empty")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Choose a project template").
				Options(options...).
				Value(&selectedTemplateID),
		),
	)

	err = form.Run()
	if err != nil {
		return nil, err
	}

	// Find the selected template config
	var selectedConfig templates.TemplateConfig
	for _, t := range availableTemplates {
		if t.ID == selectedTemplateID {
			selectedConfig = t
			break
		}
	}

	return &PromptResult{
		ProjectName: projectName,
		Template:    selectedConfig,
	}, nil
}

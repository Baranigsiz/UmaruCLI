package prompts

import (
	"fmt"
	"strings"
	"umaru/internal/config"
	"umaru/internal/generator"
	"umaru/internal/templates"

	"github.com/charmbracelet/huh"
)

type PromptResult struct {
	ProjectName    string
	Template       templates.TemplateConfig
	PackageManager string
	Addons         generator.AddonConfig
}

func Run(initialName string, initialTemplateID string, initialPkgManager string, initialAddons generator.AddonConfig, skipAddons bool) (*PromptResult, error) {
	projectName := strings.TrimSpace(initialName)
	selectedTemplateID := strings.TrimSpace(initialTemplateID)
	selectedPkgManager := strings.TrimSpace(initialPkgManager)
	selectedAddons := initialAddons

	// Load persistent user config
	userCfg := config.LoadUserConfig()
	if selectedPkgManager == "" && userCfg.PackageManager != "" {
		selectedPkgManager = userCfg.PackageManager
	}

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

	// If non-interactive full arguments are provided, return immediately
	if projectName != "" && preSelectedTmpl != nil {
		if !preSelectedTmpl.IsNodeBased() || selectedPkgManager != "" {
			return &PromptResult{
				ProjectName:    projectName,
				Template:       *preSelectedTmpl,
				PackageManager: selectedPkgManager,
				Addons:         selectedAddons,
			}, nil
		}
	}

	// 1. First Group: Project Name & Category Selection
	selectedCategory := "all"
	var firstFields []huh.Field

	if projectName == "" {
		firstFields = append(firstFields,
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
		categoryOptions := []huh.Option[string]{
			huh.NewOption("🌟 All Templates (Show all 12 starters)", "all"),
			huh.NewOption("🌐 Frontend Frameworks (React, Vue 3, Next.js, Astro)", "Frontend"),
			huh.NewOption("⚙️ Backend APIs (Go, NestJS, Express, FastAPI, Rust)", "Backend"),
			huh.NewOption("📦 Fullstack Monorepos (Go Fiber + React Vite)", "Fullstack"),
		}

		firstFields = append(firstFields,
			huh.NewSelect[string]().
				Title("Select a category").
				Options(categoryOptions...).
				Value(&selectedCategory),
		)
	}

	if len(firstFields) > 0 {
		form := huh.NewForm(huh.NewGroup(firstFields...))
		if err := form.Run(); err != nil {
			return nil, err
		}
	}

	// 2. Second Group: Template Selection based on category
	if selectedTemplateID == "" {
		var filteredTemplates []templates.TemplateConfig
		for _, t := range availableTemplates {
			if selectedCategory == "all" || t.GetCategory() == selectedCategory {
				filteredTemplates = append(filteredTemplates, t)
			}
		}

		var options []huh.Option[string]
		for _, t := range filteredTemplates {
			options = append(options, huh.NewOption(fmt.Sprintf("%s - %s", t.Name, t.Description), t.ID))
		}

		tmplForm := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a starter template").
				Options(options...).
				Value(&selectedTemplateID),
		))

		if err := tmplForm.Run(); err != nil {
			return nil, err
		}
	}

	// Resolve the selected template config
	tmpl, err := templates.FindTemplateByID(selectedTemplateID)
	if err != nil {
		return nil, err
	}

	// 3. Third Group: If template is Node-based and package manager is not specified, ask for it
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

	// 4. Fourth Group: Interactive Addon Wizard (for Backend & Fullstack templates)
	category := tmpl.GetCategory()
	if !skipAddons && (category == "Backend" || category == "Fullstack") {
		selectedDB := "none"
		selectedAuth := "none"
		var enableRedis bool

		dbOptions := []huh.Option[string]{
			huh.NewOption("None (Skip database setup)", "none"),
			huh.NewOption("PostgreSQL (Production-ready relational DB)", "postgres"),
			huh.NewOption("SQLite (Lightweight file-based embedded DB)", "sqlite"),
		}

		authOptions := []huh.Option[string]{
			huh.NewOption("None (Public API / Custom Auth)", "none"),
			huh.NewOption("JWT (JSON Web Token authentication)", "jwt"),
		}

		addonForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose a database addon (Optional)").
					Options(dbOptions...).
					Value(&selectedDB),
				huh.NewSelect[string]().
					Title("Choose an authentication addon (Optional)").
					Options(authOptions...).
					Value(&selectedAuth),
				huh.NewConfirm().
					Title("Include Redis cache support?").
					Value(&enableRedis),
			),
		)

		if err := addonForm.Run(); err != nil {
			return nil, err
		}

		selectedAddons.Database = selectedDB
		selectedAddons.Auth = selectedAuth
		selectedAddons.Redis = enableRedis
	}

	return &PromptResult{
		ProjectName:    strings.TrimSpace(projectName),
		Template:       *tmpl,
		PackageManager: selectedPkgManager,
		Addons:         selectedAddons,
	}, nil
}

package prompts

import (
	"fmt"
	"os"
	"strings"
	"umaru/internal/config"
	"umaru/internal/generator"
	"umaru/internal/templates"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
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
		if selectedPkgManager == "" && len(tmpl.InstallCommand) > 0 && tmpl.InstallCommand[0] == "bun" {
			selectedPkgManager = "bun"
		}
	}

	// If non-interactive full arguments are provided, return immediately
	if projectName != "" && preSelectedTmpl != nil {
		nodeSatisfied := !preSelectedTmpl.IsNodeBased() || selectedPkgManager != ""
		addonsSatisfied := skipAddons || initialAddons.HasAddons() || !generator.TemplateSupportsAddons(preSelectedTmpl.ID)

		if nodeSatisfied && addonsSatisfied {
			return &PromptResult{
				ProjectName:    projectName,
				Template:       *preSelectedTmpl,
				PackageManager: selectedPkgManager,
				Addons:         selectedAddons,
			}, nil
		}
	}

	// If standard input is not a terminal, we cannot prompt interactively
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return nil, fmt.Errorf("interactive prompt unavailable: standard input is not a terminal. Please provide project name and required flags (e.g. --template, --package-manager) or use --yes (-y)")
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
			huh.NewOption("🔍 Search templates by keyword...", "search"),
			huh.NewOption(fmt.Sprintf("🌟 All Templates (Show all %d starters)", len(availableTemplates)), "all"),
			huh.NewOption("🌐 Frontend Frameworks (React, Vue 3, Svelte 5, Next.js, Astro)", "Frontend"),
			huh.NewOption("⚙️ Backend APIs (Go, NestJS, Express, Hono, Fastify, Echo, FastAPI, Rust)", "Backend"),
			huh.NewOption("📦 Fullstack Monorepos (Go + React, TypeScript Monorepo)", "Fullstack"),
			huh.NewOption("⚡ CLI & Terminal Tools (Go Cobra, Bubble Tea, Lipgloss)", "CLI"),
			huh.NewOption("🖥️ Desktop Applications (Tauri v2, React, Rust)", "Desktop"),
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

	// 2. Second Group: Template Selection based on category or search
	if selectedTemplateID == "" {
		var filteredTemplates []templates.TemplateConfig
		if selectedCategory == "search" {
			var searchQuery string
			searchForm := huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Enter search term").
					Description("Matches name, ID, framework, or description (e.g. fiber, react, axum)").
					Value(&searchQuery),
			))
			if err := searchForm.Run(); err != nil {
				return nil, err
			}
			filteredTemplates = FilterTemplatesByKeyword(availableTemplates, searchQuery)
			if len(filteredTemplates) == 0 {
				fmt.Printf("\n⚠️ No templates matched '%s'. Showing all available templates.\n\n", searchQuery)
				filteredTemplates = availableTemplates
			}
		} else {
			for _, t := range availableTemplates {
				if selectedCategory == "all" || t.GetCategory() == selectedCategory {
					filteredTemplates = append(filteredTemplates, t)
				}
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
		if len(tmpl.InstallCommand) > 0 && tmpl.InstallCommand[0] == "bun" {
			selectedPkgManager = "bun"
		} else {
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
	}

	// 4. Fourth Group: Interactive Addon Wizard (for templates supporting addons)
	if !skipAddons && generator.TemplateSupportsAddons(tmpl.ID) {
		selectedDB := selectedAddons.Database
		if selectedDB == "" {
			selectedDB = "none"
		}
		selectedAuth := selectedAddons.Auth
		if selectedAuth == "" {
			selectedAuth = "none"
		}
		enableRedis := selectedAddons.Redis
		enableDocker := selectedAddons.Docker
		enableCI := selectedAddons.CI

		isRust := strings.HasPrefix(tmpl.ID, "rust-")
		var addonFields []huh.Field

		if !isRust {
			dbOptions := []huh.Option[string]{
				huh.NewOption("None (Skip database setup)", "none"),
				huh.NewOption("PostgreSQL (Production-ready relational DB)", "postgres"),
				huh.NewOption("SQLite (Lightweight file-based embedded DB)", "sqlite"),
			}

			authOptions := []huh.Option[string]{
				huh.NewOption("None (Public API / Custom Auth)", "none"),
				huh.NewOption("JWT (JSON Web Token authentication)", "jwt"),
			}

			addonFields = append(addonFields,
				huh.NewSelect[string]().
					Title("Choose a database addon (Optional)").
					Options(dbOptions...).
					Value(&selectedDB),
				huh.NewSelect[string]().
					Title("Choose an authentication addon (Optional)").
					Options(authOptions...).
					Value(&selectedAuth),
			)
		}

		addonFields = append(addonFields,
			huh.NewConfirm().
				Title("Include Redis cache support?").
				Value(&enableRedis),
			huh.NewConfirm().
				Title("Include Docker & Compose containerization?").
				Value(&enableDocker),
			huh.NewConfirm().
				Title("Include GitHub Actions CI workflow?").
				Value(&enableCI),
		)

		addonForm := huh.NewForm(huh.NewGroup(addonFields...))

		if err := addonForm.Run(); err != nil {
			return nil, err
		}

		selectedAddons.Database = selectedDB
		selectedAddons.Auth = selectedAuth
		selectedAddons.Redis = enableRedis
		selectedAddons.Docker = enableDocker
		selectedAddons.CI = enableCI
	}

	return &PromptResult{
		ProjectName:    strings.TrimSpace(projectName),
		Template:       *tmpl,
		PackageManager: selectedPkgManager,
		Addons:         selectedAddons,
	}, nil
}

// FilterTemplatesByKeyword filters a slice of templates matching any keyword in ID, Name, Description, or Category
func FilterTemplatesByKeyword(all []templates.TemplateConfig, query string) []templates.TemplateConfig {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}

	var results []templates.TemplateConfig
	for _, t := range all {
		if strings.Contains(strings.ToLower(t.ID), q) ||
			strings.Contains(strings.ToLower(t.Name), q) ||
			strings.Contains(strings.ToLower(t.Description), q) ||
			strings.Contains(strings.ToLower(t.GetCategory()), q) {
			results = append(results, t)
		}
	}
	return results
}


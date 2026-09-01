package templates

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
)

type TemplateConfig struct {
	ID             string   `json:"-"` // This will be the directory name (e.g. "go-fiber")
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Category       string   `json:"category,omitempty"`
	InstallCommand []string `json:"installCommand"`
	RunCommand     string   `json:"runCommand"`
}

// GetCategory returns the category of the template (Frontend, Backend, Fullstack)
func (t TemplateConfig) GetCategory() string {
	if t.Category != "" {
		return t.Category
	}
	switch t.ID {
	case "react-vite-ts", "vue-vite-ts", "nextjs-tailwind", "astro-tailwind":
		return "Frontend"
	case "fullstack-go-react":
		return "Fullstack"
	default:
		return "Backend"
	}
}

// GetAvailableTemplates scans the embedded FS for template.json files
func GetAvailableTemplates() ([]TemplateConfig, error) {
	var templates []TemplateConfig

	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Try to read template.json inside this directory
		configPath := path.Join(entry.Name(), "template.json")
		content, err := FS.ReadFile(configPath)
		if err != nil {
			// If template.json doesn't exist, skip this directory
			continue
		}

		var config TemplateConfig
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, err
		}
		config.ID = entry.Name()

		templates = append(templates, config)
	}

	return templates, nil
}

// FindTemplateByID looks for a template matching the given ID.
func FindTemplateByID(id string) (*TemplateConfig, error) {
	templates, err := GetAvailableTemplates()
	if err != nil {
		return nil, err
	}

	for _, t := range templates {
		if t.ID == id {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("template '%s' not found. Run 'umaru list' to view available templates", id)
}

// IsNodeBased checks if the template uses a Node.js-based package manager
func (t TemplateConfig) IsNodeBased() bool {
	if len(t.InstallCommand) == 0 {
		return false
	}
	cmd := t.InstallCommand[0]
	return cmd == "npm" || cmd == "pnpm" || cmd == "yarn" || cmd == "bun"
}

// GetInstallCommand returns the install command tailored for the specified package manager
func (t TemplateConfig) GetInstallCommand(pkgManager string) []string {
	if !t.IsNodeBased() || pkgManager == "" {
		return t.InstallCommand
	}
	switch pkgManager {
	case "pnpm":
		return []string{"pnpm", "install"}
	case "yarn":
		return []string{"yarn", "install"}
	case "bun":
		return []string{"bun", "install"}
	default:
		return []string{"npm", "install"}
	}
}

// GetRunCommand returns the run command tailored for the specified package manager
func (t TemplateConfig) GetRunCommand(pkgManager string) string {
	if !t.IsNodeBased() || pkgManager == "" {
		return t.RunCommand
	}
	switch pkgManager {
	case "pnpm":
		return "pnpm dev"
	case "yarn":
		return "yarn dev"
	case "bun":
		return "bun run dev"
	default:
		return "npm run dev"
	}
}


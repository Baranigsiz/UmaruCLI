package templates

import (
	"encoding/json"
	"io/fs"
	"path"
)

type TemplateConfig struct {
	ID             string   `json:"-"` // This will be the directory name (e.g. "go-fiber")
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	InstallCommand []string `json:"installCommand"`
	RunCommand     string   `json:"runCommand"`
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

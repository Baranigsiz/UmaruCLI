package generator

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"umaru/internal/templates"
)

type ProjectConfig struct {
	ProjectName string
	Template    string
}

func Generate(config ProjectConfig) error {

	err := fs.WalkDir(templates.FS, config.Template, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Calculate relative path to mirror the directory structure
		relPath, err := filepath.Rel(config.Template, path)
		if err != nil {
			return err
		}

		// Remove .tmpl extension if exists
		destPath := filepath.Join(config.ProjectName, relPath)
		if strings.HasSuffix(destPath, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
		}

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Read template content
		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse and execute template
		tmpl, err := template.New(destPath).Parse(string(content))
		if err != nil {
			return err
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		if err := tmpl.Execute(destFile, config); err != nil {
			return err
		}

		return nil
	})

	return err
}

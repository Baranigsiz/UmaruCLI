package generator

import (
	"fmt"
	"io"
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

// CheckDestination checks if the destination directory exists and whether it is empty.
// If force is true, an existing non-empty directory is permitted.
func CheckDestination(destDir string, force bool) error {
	info, err := os.Stat(destDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("target '%s' exists and is not a directory", destDir)
	}

	if force {
		return nil
	}

	f, err := os.Open(destDir)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read one entry to check if empty
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return nil // directory is empty
	}
	if err == nil {
		return fmt.Errorf("directory '%s' already exists and is not empty. Use --force to proceed anyway", destDir)
	}

	return err
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

		// Do not copy internal template.json into generated projects
		if relPath == "template.json" {
			return nil
		}

		isTemplate := strings.HasSuffix(relPath, ".tmpl")

		// Remove .tmpl extension if exists
		destPath := filepath.Join(config.ProjectName, relPath)
		if isTemplate {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
		}

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Read file content
		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}

		if isTemplate {
			// Parse and execute template for .tmpl files
			tmpl, err := template.New(filepath.Base(destPath)).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			if err := tmpl.Execute(destFile, config); err != nil {
				return fmt.Errorf("failed to execute template %s: %w", path, err)
			}
		} else {
			// Write raw file bytes without template parsing (preserves binary & non-template syntax)
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

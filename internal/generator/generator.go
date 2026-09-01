package generator

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"umaru/internal/templates"
)

type ProjectConfig struct {
	ProjectName string      // Human readable or base project name
	SafeName    string      // Lowercase slug for npm, cargo, etc.
	ModuleName  string      // Safe identifier for Go modules
	TargetDir   string      // Filesystem target directory
	Template    string      // Template ID (e.g. "go-fiber")
	Author      string      // Project author name
	License     string      // Project license (e.g. "MIT")
	Addons      AddonConfig // Optional Addon modules (db, auth, redis)
}

// Transliterate replaces Unicode special/accented characters with ASCII equivalents
func Transliterate(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case 'ç', 'Ç':
			sb.WriteString("c")
		case 'ğ', 'Ğ':
			sb.WriteString("g")
		case 'ı', 'İ', 'I':
			sb.WriteString("i")
		case 'ö', 'Ö':
			sb.WriteString("o")
		case 'ş', 'Ş':
			sb.WriteString("s")
		case 'ü', 'Ü':
			sb.WriteString("u")
		case 'á', 'à', 'ä', 'â', 'ã', 'å', 'Á', 'À', 'Ä', 'Â', 'Ã', 'Å':
			sb.WriteString("a")
		case 'é', 'è', 'ë', 'ê', 'É', 'È', 'Ë', 'Ê':
			sb.WriteString("e")
		case 'í', 'ì', 'ï', 'î', 'Í', 'Ì', 'Ï', 'Î':
			sb.WriteString("i")
		case 'ó', 'ò', 'ô', 'õ', 'Ó', 'Ò', 'Ô', 'Õ':
			sb.WriteString("o")
		case 'ú', 'ù', 'û', 'Ú', 'Ù', 'Û':
			sb.WriteString("u")
		case 'ñ', 'Ñ':
			sb.WriteString("n")
		case 'ß':
			sb.WriteString("ss")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Slugify converts any string into a clean lowercase slug (e.g. "Türkçe Proje" -> "turkce-proje")
func Slugify(s string) string {
	s = Transliterate(s)
	s = strings.TrimSpace(strings.ToLower(s))
	// Replace non-alphanumeric characters (excluding hyphen and underscore) with hyphen
	reg := regexp.MustCompile(`[^a-z0-9_\-]+`)
	s = reg.ReplaceAllString(s, "-")
	// Trim leading and trailing hyphens
	s = strings.Trim(s, "-")
	if s == "" {
		return "umaru-app"
	}
	return s
}

// ResolveProjectConfig determines the target directory and normalized project names
func ResolveProjectConfig(rawInput string, templateID string) (ProjectConfig, error) {
	rawInput = strings.TrimSpace(rawInput)
	if rawInput == "" {
		rawInput = "."
	}

	targetDir := filepath.Clean(rawInput)
	var projectName string

	if targetDir == "." {
		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			return ProjectConfig{}, fmt.Errorf("failed to resolve current directory: %w", err)
		}
		projectName = filepath.Base(absPath)
	} else {
		projectName = filepath.Base(targetDir)
	}

	safeName := Slugify(projectName)
	moduleName := safeName

	return ProjectConfig{
		ProjectName: projectName,
		SafeName:    safeName,
		ModuleName:  moduleName,
		TargetDir:   targetDir,
		Template:    templateID,
		License:     "MIT",
	}, nil
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

// DryRun returns the list of file paths that would be generated without writing to disk
func DryRun(config ProjectConfig) ([]string, error) {
	var files []string
	err := fs.WalkDir(templates.FS, config.Template, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(config.Template, path)
		if err != nil {
			return err
		}

		if relPath == "template.json" {
			return nil
		}

		destPath := filepath.Join(config.TargetDir, relPath)
		destPath = strings.TrimSuffix(destPath, ".tmpl")

		files = append(files, destPath)
		return nil
	})

	// Append Addon files if selected
	addonFiles := GetAddonFiles(config)
	files = append(files, addonFiles...)

	return files, err
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
		destPath := filepath.Join(config.TargetDir, relPath)
		destPath = strings.TrimSuffix(destPath, ".tmpl")

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Read file content
		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}

		// Choose proper file permission: 0755 for shell scripts, 0644 for regular files
		var perm os.FileMode = 0644
		if strings.HasSuffix(destPath, ".sh") {
			perm = 0755
		}

		if isTemplate {
			// Parse and execute template for .tmpl files
			tmpl, err := template.New(filepath.Base(destPath)).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}

			destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
			if err != nil {
				return err
			}

			execErr := tmpl.Execute(destFile, config)
			closeErr := destFile.Close()
			if execErr != nil {
				return fmt.Errorf("failed to execute template %s: %w", path, execErr)
			}
			if closeErr != nil {
				return closeErr
			}
		} else {
			// Write raw file bytes without template parsing (preserves binary & non-template syntax)
			if err := os.WriteFile(destPath, content, perm); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Generate selected Addons
	return GenerateAddons(config)
}

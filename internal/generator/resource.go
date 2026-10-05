package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ResourceConfig defines the parameters for generating a fullstack/backend resource
type ResourceConfig struct {
	Name      string // Raw input (e.g. "Product", "user_profile")
	TargetDir string // Project root directory
	Force     bool   // Overwrite existing files
	DryRun    bool   // Simulate without writing
}

// ResourceFile describes a single generated file
type ResourceFile struct {
	Path    string `json:"path"`
	RelPath string `json:"rel_path"`
	Content string `json:"content"`
	Action  string `json:"action"` // "created", "skipped (already exists)", "overwritten"
}

// ResourceResult contains the report of resource generation
type ResourceResult struct {
	ResourceName string         `json:"resource_name"`
	Language     string         `json:"language"`
	Framework    string         `json:"framework"`
	Files        []ResourceFile `json:"files"`
}

type resourceContext struct {
	pascal       string
	slug         string
	camel        string
	pluralSlug   string
	pluralPascal string
	framework    string
	moduleName   string // Go module name from go.mod (e.g. "github.com/user/my-api")
}

// ToPascalCase converts strings like "user_profile", "user-profile", "user" into "UserProfile"
func ToPascalCase(s string) string {
	s = Transliterate(s)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})
	if len(parts) == 0 {
		return "Resource"
	}
	var sb strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		sb.WriteRune(unicode.ToUpper(runes[0]))
		for i := 1; i < len(runes); i++ {
			sb.WriteRune(unicode.ToLower(runes[i]))
		}
	}
	res := sb.String()
	if res == "" {
		return "Resource"
	}
	return res
}

// ToCamelCase converts string into camelCase (e.g. "UserProfile" -> "userProfile")
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return "resource"
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Pluralize adds simple English pluralization (e.g. "user" -> "users", "category" -> "categories")
func Pluralize(s string) string {
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		lastTwo := lower[len(lower)-2:]
		if !strings.ContainsAny(string(lastTwo[0]), "aeiou") {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}

// GenerateResource inspects project framework and generates model, handler, service, and repository boilerplate
func GenerateResource(cfg ResourceConfig) (*ResourceResult, error) {
	targetDir := cfg.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory: %w", err)
	}

	proj, err := DetectProject(absDir)
	if err != nil {
		return nil, fmt.Errorf("could not detect project structure: %w", err)
	}

	rawName := strings.TrimSpace(cfg.Name)
	if rawName == "" {
		return nil, fmt.Errorf("resource name cannot be empty")
	}

	rCtx := resourceContext{
		pascal:       ToPascalCase(rawName),
		slug:         Slugify(rawName),
		camel:        ToCamelCase(rawName),
		pluralSlug:   Pluralize(Slugify(rawName)),
		pluralPascal: Pluralize(ToPascalCase(rawName)),
		framework:    proj.Framework,
		moduleName:   proj.ModuleName,
	}

	baseDir := GetAddonBaseDir(proj.ToProjectConfig(AddonConfig{}))

	var files []ResourceFile

	switch proj.Type {
	case ProjectTypeGo:
		files = generateGoResource(baseDir, rCtx)
	case ProjectTypeNode:
		files = generateNodeResource(baseDir, rCtx)
	case ProjectTypePython:
		files = generatePythonResource(baseDir, rCtx)
	case ProjectTypeRust:
		files = generateRustResource(baseDir, rCtx)
	default:
		return nil, fmt.Errorf("resource generation is not supported for project type: %s", proj.Type)
	}

	result := &ResourceResult{
		ResourceName: rCtx.pascal,
		Language:     string(proj.Type),
		Framework:    proj.Framework,
		Files:        make([]ResourceFile, 0, len(files)),
	}

	for _, f := range files {
		action := "created"
		if fileExists(f.Path) {
			if !cfg.Force {
				action = "skipped (already exists)"
				f.Action = action
				result.Files = append(result.Files, f)
				continue
			}
			action = "overwritten"
		}

		if !cfg.DryRun {
			if err := os.MkdirAll(filepath.Dir(f.Path), 0755); err != nil {
				return nil, fmt.Errorf("failed creating directory for %s: %w", f.RelPath, err)
			}
			if err := os.WriteFile(f.Path, []byte(f.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed writing %s: %w", f.RelPath, err)
			}
		}
		f.Action = action
		result.Files = append(result.Files, f)
	}

	return result, nil
}

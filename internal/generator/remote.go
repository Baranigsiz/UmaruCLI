package generator

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"umaru/internal/templates"
)

// NormalizeGitURL converts GitHub shorthands (e.g., "owner/repo") to full git clone URLs
func NormalizeGitURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// If already a full URL or SSH
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "git@") {
		return raw
	}

	// Shorthand "owner/repo"
	parts := strings.Split(raw, "/")
	if len(parts) == 2 && !strings.Contains(raw, ":") {
		return fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1])
	}

	return raw
}

// GenerateFromRemote clones a remote repository into the target directory,
// strips the .git metadata, and renders any .tmpl files.
func GenerateFromRemote(repoURL string, config ProjectConfig) (*templates.TemplateConfig, error) {
	normalizedURL := NormalizeGitURL(repoURL)
	if normalizedURL == "" {
		return nil, fmt.Errorf("invalid or empty remote repository URL")
	}

	// Ensure destination directory is created
	if err := os.MkdirAll(config.TargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory %s: %w", config.TargetDir, err)
	}

	// Clone repo with depth 1
	cloneCmd := exec.Command("git", "clone", "--depth", "1", normalizedURL, config.TargetDir)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return nil, fmt.Errorf("git clone failed: %s", outStr)
		}
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	// Remove existing .git directory
	gitDir := filepath.Join(config.TargetDir, ".git")
	_ = os.RemoveAll(gitDir)

	// Check if template.json exists in remote repository
	var templateConfig templates.TemplateConfig
	remoteTemplateJSON := filepath.Join(config.TargetDir, "template.json")
	if data, err := os.ReadFile(remoteTemplateJSON); err == nil {
		_ = json.Unmarshal(data, &templateConfig)
		_ = os.Remove(remoteTemplateJSON) // remove template.json from output
	}

	if templateConfig.Name == "" {
		templateConfig.Name = filepath.Base(normalizedURL)
	}

	// Walk and process any .tmpl files
	err := filepath.WalkDir(config.TargetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".tmpl") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			destPath := strings.TrimSuffix(path, ".tmpl")

			tmpl, err := template.New(filepath.Base(destPath)).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse remote template file %s: %w", path, err)
			}

			destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}

			execErr := tmpl.Execute(destFile, config)
			closeErr := destFile.Close()
			if execErr != nil {
				return execErr
			}
			if closeErr != nil {
				return closeErr
			}

			// Delete the original .tmpl file
			_ = os.Remove(path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed processing remote template files: %w", err)
	}

	return &templateConfig, nil
}

// DryRunRemote clones to a temporary directory to simulate generated files
func DryRunRemote(repoURL string, config ProjectConfig) ([]string, error) {
	tempDir, err := os.MkdirTemp("", "umaru-remote-dryrun-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	normalizedURL := NormalizeGitURL(repoURL)
	cloneCmd := exec.Command("git", "clone", "--depth", "1", normalizedURL, tempDir)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("remote dry-run clone failed: %s", string(out))
	}

	var files []string
	err = filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(relPath, ".git") || relPath == "template.json" {
			return nil
		}

		destPath := filepath.Join(config.TargetDir, relPath)
		destPath = strings.TrimSuffix(destPath, ".tmpl")

		files = append(files, destPath)
		return nil
	})

	return files, err
}

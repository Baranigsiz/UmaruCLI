package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"umaru/internal/templates"
)

// CloneTimeout is the maximum duration allowed for a git clone operation
const CloneTimeout = 2 * time.Minute

// RemoteSpec contains the parsed clone URL and optional git ref (branch/tag/commit)
type RemoteSpec struct {
	CloneURL string
	Ref      string
}

// ParseRemoteURL parses a git shorthand, full URL, or branch/tag specifier into a RemoteSpec.
// Supported formats:
//   - "owner/repo" -> https://github.com/owner/repo.git
//   - "owner/repo#dev" -> https://github.com/owner/repo.git, ref: "dev"
//   - "github.com/owner/repo" -> https://github.com/owner/repo.git
//   - "gitlab.com/owner/repo#v1.0.0" -> https://gitlab.com/owner/repo.git, ref: "v1.0.0"
//   - "github:owner/repo" -> https://github.com/owner/repo.git
//   - "gitlab:owner/repo" -> https://gitlab.com/owner/repo.git
//   - "https://github.com/owner/repo.git#main" -> https://github.com/owner/repo.git, ref: "main"
//   - "git@github.com:owner/repo.git#main" -> git@github.com:owner/repo.git, ref: "main"
func ParseRemoteURL(raw string) (*RemoteSpec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "-") {
		return nil, fmt.Errorf("invalid or empty remote repository URL")
	}

	var ref string
	if idx := strings.Index(raw, "#"); idx != -1 {
		ref = strings.TrimSpace(raw[idx+1:])
		raw = strings.TrimSpace(raw[:idx])
	}

	if raw == "" || strings.HasPrefix(raw, "-") {
		return nil, fmt.Errorf("invalid repository URL")
	}

	// Handle shorthand prefixes like github:owner/repo, gh:owner/repo, gitlab:owner/repo
	if strings.HasPrefix(raw, "github:") || strings.HasPrefix(raw, "gh:") {
		raw = strings.TrimPrefix(strings.TrimPrefix(raw, "github:"), "gh:")
		raw = "https://github.com/" + strings.TrimPrefix(raw, "/") + ".git"
	} else if strings.HasPrefix(raw, "gitlab:") {
		raw = "https://gitlab.com/" + strings.TrimPrefix(strings.TrimPrefix(raw, "gitlab:"), "/") + ".git"
	} else if strings.HasPrefix(raw, "github.com/") {
		raw = "https://" + raw
		if !strings.HasSuffix(raw, ".git") {
			raw += ".git"
		}
	} else if strings.HasPrefix(raw, "gitlab.com/") {
		raw = "https://" + raw
		if !strings.HasSuffix(raw, ".git") {
			raw += ".git"
		}
	} else if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") && !strings.HasPrefix(raw, "git@") {
		parts := strings.Split(raw, "/")
		if len(parts) == 2 && !strings.Contains(raw, ":") {
			raw = fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1])
		}
	}

	return &RemoteSpec{
		CloneURL: raw,
		Ref:      ref,
	}, nil
}

// NormalizeGitURL retains backwards compatibility for callers expecting just the URL string
func NormalizeGitURL(raw string) string {
	spec, err := ParseRemoteURL(raw)
	if err != nil {
		return ""
	}
	return spec.CloneURL
}

// GenerateFromRemote clones a remote repository into a temporary directory,
// strips the .git metadata, renders any .tmpl files, and copies the result
// into the target directory.
func GenerateFromRemote(repoURL string, config ProjectConfig) (*templates.TemplateConfig, error) {
	spec, err := ParseRemoteURL(repoURL)
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp("", "umaru-remote-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary clone directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repo with depth 1 into temporary directory
	ctx, cancel := context.WithTimeout(context.Background(), CloneTimeout)
	defer cancel()

	var cloneArgs []string
	if spec.Ref != "" {
		cloneArgs = []string{"clone", "--depth", "1", "--branch", spec.Ref, spec.CloneURL, tempDir}
	} else {
		cloneArgs = []string{"clone", "--depth", "1", spec.CloneURL, tempDir}
	}

	cloneCmd := exec.CommandContext(ctx, "git", cloneArgs...)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git clone timed out after %s — check your network or repository URL", CloneTimeout)
		}
		outStr := strings.TrimSpace(string(out))
		if outStr != "" {
			return nil, fmt.Errorf("git clone failed: %s", outStr)
		}
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	// Remove existing .git directory
	gitDir := filepath.Join(tempDir, ".git")
	_ = os.RemoveAll(gitDir)

	// Check if template.json exists in remote repository
	var templateConfig templates.TemplateConfig
	remoteTemplateJSON := filepath.Join(tempDir, "template.json")
	if data, err := os.ReadFile(remoteTemplateJSON); err == nil {
		_ = json.Unmarshal(data, &templateConfig)
		_ = os.Remove(remoteTemplateJSON) // remove template.json from output
	}

	if templateConfig.Name == "" {
		templateConfig.Name = filepath.Base(spec.CloneURL)
	}

	// Walk and process any .tmpl files
	err = filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
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

			var perm os.FileMode = 0644
			if info, err := d.Info(); err == nil && (info.Mode()&0111 != 0) {
				perm = 0755
			} else if strings.HasSuffix(destPath, ".sh") {
				perm = 0755
			}

			destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
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

	// Ensure destination directory is created
	if err := os.MkdirAll(config.TargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory %s: %w", config.TargetDir, err)
	}

	// Copy all files and folders from tempDir to config.TargetDir
	err = filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(config.TargetDir, relPath)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		info, err := d.Info()
		perm := fs.FileMode(0644)
		if err == nil {
			perm = info.Mode().Perm()
		}

		return copyFile(path, targetPath, perm)
	})

	if err != nil {
		return nil, fmt.Errorf("failed copying files to target directory: %w", err)
	}

	return &templateConfig, nil
}

func copyFile(srcPath, dstPath string, perm fs.FileMode) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

// DryRunRemote clones to a temporary directory to simulate generated files
func DryRunRemote(repoURL string, config ProjectConfig) ([]string, error) {
	spec, err := ParseRemoteURL(repoURL)
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp("", "umaru-remote-dryrun-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), CloneTimeout)
	defer cancel()

	var cloneArgs []string
	if spec.Ref != "" {
		cloneArgs = []string{"clone", "--depth", "1", "--branch", spec.Ref, spec.CloneURL, tempDir}
	} else {
		cloneArgs = []string{"clone", "--depth", "1", spec.CloneURL, tempDir}
	}

	cloneCmd := exec.CommandContext(ctx, "git", cloneArgs...)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git clone timed out after %s — check your network or repository URL", CloneTimeout)
		}
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

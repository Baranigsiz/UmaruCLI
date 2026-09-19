package cleaner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ArtifactCategory string

const (
	CategoryNode   ArtifactCategory = "Node.js"
	CategoryBuild  ArtifactCategory = "Build Output"
	CategoryCache  ArtifactCategory = "Cache"
	CategoryRust   ArtifactCategory = "Rust"
	CategoryPython ArtifactCategory = "Python"
	CategoryOS     ArtifactCategory = "OS / Temp"
)

// TargetItem represents a single removable file or directory
type TargetItem struct {
	Path        string           `json:"path"`
	RelPath     string           `json:"rel_path"`
	Name        string           `json:"name"`
	Category    ArtifactCategory `json:"category"`
	SizeBytes   int64            `json:"size_bytes"`
	SizeDisplay string           `json:"size_display"`
	IsDir       bool             `json:"is_dir"`
}

// ScanOptions configures the scanner
type ScanOptions struct {
	RootDir    string
	Recursive  bool
	IncludeAll bool // If true, also includes virtual environments like .venv
}

// CleanReport represents the summary of detected removable items
type CleanReport struct {
	RootDir        string       `json:"root_dir"`
	Items          []TargetItem `json:"items"`
	TotalSizeBytes int64        `json:"total_size_bytes"`
	TotalDisplay   string       `json:"total_size_display"`
}

// CleanResult holds the outcome of an executed cleanup
type CleanResult struct {
	DeletedCount     int    `json:"deleted_count"`
	ReclaimedBytes   int64  `json:"reclaimed_bytes"`
	ReclaimedDisplay string `json:"reclaimed_display"`
}

// FormatBytes formats byte counts into human-readable strings (e.g. "12.4 MB")
func FormatBytes(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), units[exp])
}

// DirSize calculates the recursive byte size of a directory
func DirSize(path string) int64 {
	var size int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size
}

// matchDirectoryTarget checks if a directory name is a cleanup candidate
func matchDirectoryTarget(name string, includeAll bool) (bool, ArtifactCategory) {
	switch strings.ToLower(name) {
	case "node_modules":
		return true, CategoryNode
	case "dist", "build", "out", "coverage":
		return true, CategoryBuild
	case ".next", ".nuxt", ".turbo", ".astro", ".svelte-kit", ".cache":
		return true, CategoryCache
	case "target":
		return true, CategoryRust
	case "__pycache__", ".pytest_cache", ".mypy_cache", ".ruff_cache":
		return true, CategoryPython
	case ".venv", "venv":
		if includeAll {
			return true, CategoryPython
		}
		return false, ""
	case "tmp":
		return true, CategoryOS
	default:
		return false, ""
	}
}

// matchFileTarget checks if an individual file is a cleanup candidate
func matchFileTarget(name string) (bool, ArtifactCategory) {
	lower := strings.ToLower(name)
	if lower == ".ds_store" || lower == "thumbs.db" {
		return true, CategoryOS
	}
	if strings.HasSuffix(lower, ".pyc") || strings.HasSuffix(lower, ".pyo") {
		return true, CategoryPython
	}
	return false, ""
}

// Scan inspects the RootDir for cleanup candidates
func Scan(opts ScanOptions) (*CleanReport, error) {
	absRoot, err := filepath.Abs(opts.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("target path error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target path is not a directory: %s", absRoot)
	}

	var items []TargetItem
	var totalBytes int64

	// If recursive is false, we only scan the top-level directory and immediate subdirs
	maxDepth := 1
	if opts.Recursive {
		maxDepth = 20
	}

	cleanRoot := filepath.Clean(absRoot)

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // ignore permission or access errors gracefully
		}

		cleanPath := filepath.Clean(path)
		if cleanPath == cleanRoot {
			return nil
		}

		name := d.Name()

		// NEVER touch or recurse into .git repository folders
		if name == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate relative depth from root
		rel, err := filepath.Rel(cleanRoot, cleanPath)
		if err != nil {
			return nil
		}
		depth := len(strings.Split(rel, string(filepath.Separator)))
		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if isTarget, category := matchDirectoryTarget(name, opts.IncludeAll); isTarget {
				size := DirSize(path)
				items = append(items, TargetItem{
					Path:        path,
					RelPath:     rel,
					Name:        name,
					Category:    category,
					SizeBytes:   size,
					SizeDisplay: FormatBytes(size),
					IsDir:       true,
				})
				totalBytes += size
				return filepath.SkipDir // Do not recurse into target directories
			}
		} else {
			if isTarget, category := matchFileTarget(name); isTarget {
				fileInfo, err := d.Info()
				var size int64
				if err == nil {
					size = fileInfo.Size()
				}
				items = append(items, TargetItem{
					Path:        path,
					RelPath:     rel,
					Name:        name,
					Category:    category,
					SizeBytes:   size,
					SizeDisplay: FormatBytes(size),
					IsDir:       false,
				})
				totalBytes += size
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort items by size descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].SizeBytes > items[j].SizeBytes
	})

	return &CleanReport{
		RootDir:        absRoot,
		Items:          items,
		TotalSizeBytes: totalBytes,
		TotalDisplay:   FormatBytes(totalBytes),
	}, nil
}

// ExecuteClean deletes all items found in the report
func ExecuteClean(report *CleanReport) (*CleanResult, error) {
	var count int
	var reclaimed int64

	for _, item := range report.Items {
		if err := os.RemoveAll(item.Path); err != nil {
			return nil, fmt.Errorf("failed to remove %s: %w", item.RelPath, err)
		}
		count++
		reclaimed += item.SizeBytes
	}

	return &CleanResult{
		DeletedCount:     count,
		ReclaimedBytes:   reclaimed,
		ReclaimedDisplay: FormatBytes(reclaimed),
	}, nil
}

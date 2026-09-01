package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	RepoOwner = "Baranigsiz"
	RepoName  = "UmaruCLI"
	APIURL    = "https://api.github.com/repos/" + RepoOwner + "/" + RepoName + "/releases/latest"
)

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type ReleaseInfo struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	PublishedAt string         `json:"published_at"`
	HTMLURL     string         `json:"html_url"`
	Body        string         `json:"body"`
	Assets      []ReleaseAsset `json:"assets"`
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// FetchLatestRelease queries the GitHub Releases API for the latest version metadata
func FetchLatestRelease() (*ReleaseInfo, error) {
	req, err := http.NewRequest("GET", APIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UmaruCLI-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for %s/%s", RepoOwner, RepoName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}

	return &release, nil
}

// CleanVersion removes 'v' prefix and whitespace from version string
func CleanVersion(v string) string {
	v = strings.TrimSpace(v)
	return strings.TrimPrefix(v, "v")
}

// IsNewerVersion returns true if latest is semantically newer than current
func IsNewerVersion(current, latest string) bool {
	curr := CleanVersion(current)
	lat := CleanVersion(latest)

	if curr == "dev" || curr == "none" || curr == "unknown" || curr == "" {
		return false
	}
	if curr == lat {
		return false
	}

	currParts := strings.Split(curr, ".")
	latParts := strings.Split(lat, ".")

	for i := 0; i < len(currParts) && i < len(latParts); i++ {
		var c, l int
		_, _ = fmt.Sscanf(currParts[i], "%d", &c)
		_, _ = fmt.Sscanf(latParts[i], "%d", &l)

		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}

	return len(latParts) > len(currParts)
}

// FindAssetForSystem finds the compatible archive asset for current OS and architecture
func (r *ReleaseInfo) FindAssetForSystem() (*ReleaseAsset, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for _, asset := range r.Assets {
		name := strings.ToLower(asset.Name)
		// Match OS and Arch in filename (e.g., umaru_1.0.0_windows_amd64.zip or umaru_linux_amd64.tar.gz)
		if strings.Contains(name, goos) && strings.Contains(name, goarch) {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no compatible release asset found for %s/%s in release %s", goos, goarch, r.TagName)
}

// DownloadAndExtractBinary downloads the archive asset and extracts the umaru binary bytes
func DownloadAndExtractBinary(assetURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UmaruCLI-Updater")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download release asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with HTTP status %d", resp.StatusCode)
	}

	archiveBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive bytes: %w", err)
	}

	binaryName := "umaru"
	if runtime.GOOS == "windows" {
		binaryName = "umaru.exe"
	}

	// If zip archive (Windows)
	if strings.HasSuffix(strings.ToLower(assetURL), ".zip") {
		zipReader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
		if err != nil {
			return nil, fmt.Errorf("failed to read zip archive: %w", err)
		}

		for _, file := range zipReader.File {
			if strings.EqualFold(filepath.Base(file.Name), binaryName) {
				rc, err := file.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("binary '%s' not found inside downloaded zip", binaryName)
	}

	// If tar.gz archive (Linux / macOS)
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress gzip archive: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar: %w", err)
		}

		if strings.EqualFold(filepath.Base(header.Name), binaryName) {
			return io.ReadAll(tarReader)
		}
	}

	return nil, fmt.Errorf("binary '%s' not found inside downloaded tar.gz", binaryName)
}

// ReplaceCurrentExecutable safely updates the current running executable
func ReplaceCurrentExecutable(newBinaryBytes []byte) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to evaluate symlinks: %w", err)
	}

	dir := filepath.Dir(execPath)
	tempFile, err := os.CreateTemp(dir, "umaru-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary binary: %w", err)
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.Write(newBinaryBytes); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to write new binary: %w", err)
	}
	tempFile.Close()

	// Set executable permissions
	if err := os.Chmod(tempPath, 0755); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to set permissions on new binary: %w", err)
	}

	// On Windows, rename existing binary to .old before moving new binary into place
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath) // remove previous backup if exists
		if err := os.Rename(execPath, oldPath); err != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("failed to replace active executable: %w", err)
		}
		if err := os.Rename(tempPath, execPath); err != nil {
			// Rollback
			_ = os.Rename(oldPath, execPath)
			return fmt.Errorf("failed to move new binary into place: %w", err)
		}
		return nil
	}

	// On Linux / macOS, atomic rename
	if err := os.Rename(tempPath, execPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	return nil
}

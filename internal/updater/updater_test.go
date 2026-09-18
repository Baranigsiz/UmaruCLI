package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.0.0", "1.0.0"},
		{"v2.3.4-beta", "2.3.4-beta"},
		{"  v0.1.0  ", "0.1.0"},
		{"1.2.3", "1.2.3"},
		{"", ""},
	}

	for _, tt := range tests {
		got := CleanVersion(tt.input)
		if got != tt.expected {
			t.Errorf("CleanVersion(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"1.0.0", "1.0.1", true},
		{"v1.0.0", "v1.1.0", true},
		{"v1.2.0", "v2.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"v1.1.0", "v1.0.9", false},
		{"dev", "v1.0.0", false},
		{"none", "v1.0.0", false},
		{"unknown", "v1.0.0", false},
		{"1.0.0", "1.0.0.1", true},
		{"1.0.0-beta", "1.0.0", true},
		{"1.0.0-rc.1", "1.0.0", true},
		{"1.0.0", "1.0.0-beta", false},
		{"1.0.0-alpha", "1.0.0-beta", true},
		{"1.0.0-beta.2", "1.0.0-beta.10", true},
		{"1.0.0-beta.10", "1.0.0-beta.2", false},
		{"1.0.0-alpha", "1.0.0-alpha.1", true},
	}

	for _, tt := range tests {
		got := IsNewerVersion(tt.current, tt.latest)
		if got != tt.expected {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.expected)
		}
	}
}

func TestFindAssetForSystem(t *testing.T) {
	release := ReleaseInfo{
		TagName: "v1.0.0",
		Assets: []ReleaseAsset{
			// Add non-archive assets (checksums, sbom) that should be skipped
			{Name: "umaru_1.0.0_windows_amd64.zip.sha256", BrowserDownloadURL: "http://example.com/win64.sha256"},
			{Name: "umaru_1.0.0_windows_amd64.sbom", BrowserDownloadURL: "http://example.com/win64.sbom"},
			{Name: "umaru_1.0.0_linux_amd64.tar.gz.sha256", BrowserDownloadURL: "http://example.com/linux64.sha256"},
			// Real archives
			{Name: "umaru_1.0.0_windows_amd64.zip", BrowserDownloadURL: "http://example.com/win64.zip"},
			{Name: "umaru_1.0.0_windows_arm64.zip", BrowserDownloadURL: "http://example.com/winarm.zip"},
			{Name: "umaru_1.0.0_linux_amd64.tar.gz", BrowserDownloadURL: "http://example.com/linux64.tar.gz"},
			{Name: "umaru_1.0.0_linux_arm64.tar.gz", BrowserDownloadURL: "http://example.com/linuxarm64.tar.gz"},
			{Name: "umaru_1.0.0_darwin_amd64.tar.gz", BrowserDownloadURL: "http://example.com/darwin64.tar.gz"},
			{Name: "umaru_1.0.0_darwin_arm64.tar.gz", BrowserDownloadURL: "http://example.com/darwinarm.tar.gz"},
		},
	}

	asset, err := release.FindAssetForSystem()
	if err != nil {
		t.Fatalf("FindAssetForSystem failed for %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
	}

	if asset == nil {
		t.Fatalf("Expected non-nil asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	if strings.HasSuffix(asset.Name, ".sha256") || strings.HasSuffix(asset.Name, ".sbom") {
		t.Errorf("FindAssetForSystem matched non-archive asset: %s", asset.Name)
	}
}

func TestCleanupOldExecutable(t *testing.T) {
	// Should run cleanly without panicking on any platform
	CleanupOldExecutable()
}

func TestDownloadAndExtractBinary_Zip(t *testing.T) {
	binaryName := "umaru"
	if runtime.GOOS == "windows" {
		binaryName = "umaru.exe"
	}

	// Create in-memory zip containing binaryName
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(binaryName)
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	expectedContent := []byte("binary-test-content-123")
	if _, err := w.Write(expectedContent); err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("Failed to close zip: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	defer ts.Close()

	data, err := DownloadAndExtractBinary(ts.URL + "/test.zip")
	if err != nil {
		t.Fatalf("DownloadAndExtractBinary failed: %v", err)
	}

	if !bytes.Equal(data, expectedContent) {
		t.Errorf("Extracted bytes %q != expected %q", string(data), string(expectedContent))
	}
}

func TestDownloadAndExtractBinary_TarGz(t *testing.T) {
	binaryName := "umaru"
	if runtime.GOOS == "windows" {
		binaryName = "umaru.exe"
	}

	// Create in-memory tar.gz containing binaryName
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	expectedContent := []byte("tar-binary-content-456")
	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0755,
		Size: int64(len(expectedContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write(expectedContent); err != nil {
		t.Fatalf("Failed to write to tar: %v", err)
	}
	_ = tw.Close()
	_ = gw.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	defer ts.Close()

	data, err := DownloadAndExtractBinary(ts.URL + "/test.tar.gz")
	if err != nil {
		t.Fatalf("DownloadAndExtractBinary failed: %v", err)
	}

	if !bytes.Equal(data, expectedContent) {
		t.Errorf("Extracted bytes %q != expected %q", string(data), string(expectedContent))
	}
}

func TestDownloadAndExtractBinary_Errors(t *testing.T) {
	t.Run("404NotFound", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		_, err := DownloadAndExtractBinary(ts.URL + "/missing.zip")
		if err == nil {
			t.Errorf("Expected error for 404, got nil")
		}
	})

	t.Run("BinaryMissingInsideZip", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, _ := zw.Create("other-file.txt")
		_, _ = w.Write([]byte("not the binary"))
		_ = zw.Close()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(buf.Bytes())
		}))
		defer ts.Close()

		_, err := DownloadAndExtractBinary(ts.URL + "/empty.zip")
		if err == nil {
			t.Errorf("Expected error when binary not inside zip, got nil")
		}
	})
}

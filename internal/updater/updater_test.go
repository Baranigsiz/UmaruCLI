package updater

import (
	"runtime"
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
			{Name: "umaru_1.0.0_windows_amd64.zip", BrowserDownloadURL: "http://example.com/win64.zip"},
			{Name: "umaru_1.0.0_windows_arm64.zip", BrowserDownloadURL: "http://example.com/winarm.zip"},
			{Name: "umaru_1.0.0_linux_amd64.tar.gz", BrowserDownloadURL: "http://example.com/linux64.tar.gz"},
			{Name: "umaru_1.0.0_linux_arm64.tar.gz", BrowserDownloadURL: "http://example.com/linuxarm.tar.gz"},
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
}

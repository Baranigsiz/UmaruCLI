package generator

import (
	"testing"
)

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Baranigsiz/UmaruCLI", "https://github.com/Baranigsiz/UmaruCLI.git"},
		{"facebook/react", "https://github.com/facebook/react.git"},
		{"https://github.com/user/repo.git", "https://github.com/user/repo.git"},
		{"git@github.com:user/repo.git", "git@github.com:user/repo.git"},
		{"http://gitlab.com/user/repo", "http://gitlab.com/user/repo"},
		{"", ""},
		{"--upload-pack=exploit", ""},
		{"-oProxyCommand=exploit", ""},
		{"--depth=1", ""},
	}

	for _, tt := range tests {
		got := NormalizeGitURL(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeGitURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDryRunRemote_InvalidURL(t *testing.T) {
	cfg := ProjectConfig{TargetDir: "tmp"}
	_, err := DryRunRemote("", cfg)
	if err == nil {
		t.Errorf("Expected error for empty remote URL, got nil")
	}

	_, errFlag := DryRunRemote("--malicious-flag", cfg)
	if errFlag == nil {
		t.Errorf("Expected error for flag-like remote URL, got nil")
	}
}

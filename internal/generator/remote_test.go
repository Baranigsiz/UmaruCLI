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
	}

	for _, tt := range tests {
		got := NormalizeGitURL(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeGitURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

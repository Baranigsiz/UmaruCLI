package generator

import (
	"os"
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

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		input       string
		expectedURL string
		expectedRef string
		expectErr   bool
	}{
		{"Baranigsiz/UmaruCLI", "https://github.com/Baranigsiz/UmaruCLI.git", "", false},
		{"Baranigsiz/UmaruCLI#main", "https://github.com/Baranigsiz/UmaruCLI.git", "main", false},
		{"Baranigsiz/UmaruCLI#v1.0.0", "https://github.com/Baranigsiz/UmaruCLI.git", "v1.0.0", false},
		{"github:user/repo", "https://github.com/user/repo.git", "", false},
		{"github:user/repo#dev", "https://github.com/user/repo.git", "dev", false},
		{"gh:user/repo#feat-1", "https://github.com/user/repo.git", "feat-1", false},
		{"gitlab:user/repo", "https://gitlab.com/user/repo.git", "", false},
		{"gitlab:user/repo#release", "https://gitlab.com/user/repo.git", "release", false},
		{"github.com/user/repo", "https://github.com/user/repo.git", "", false},
		{"gitlab.com/user/repo#main", "https://gitlab.com/user/repo.git", "main", false},
		{"https://github.com/user/repo.git#dev", "https://github.com/user/repo.git", "dev", false},
		{"", "", "", true},
		{"-malicious", "", "", true},
		{"#only-ref", "", "", true},
	}

	for _, tt := range tests {
		spec, err := ParseRemoteURL(tt.input)
		if tt.expectErr {
			if err == nil {
				t.Errorf("ParseRemoteURL(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRemoteURL(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if spec.CloneURL != tt.expectedURL {
			t.Errorf("ParseRemoteURL(%q) CloneURL = %q, want %q", tt.input, spec.CloneURL, tt.expectedURL)
		}
		if spec.Ref != tt.expectedRef {
			t.Errorf("ParseRemoteURL(%q) Ref = %q, want %q", tt.input, spec.Ref, tt.expectedRef)
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

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	src := tempDir + "/src.txt"
	dst := tempDir + "/dst.txt"

	content := []byte("Hello Umaru remote copy!")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("Failed to write src file: %v", err)
	}

	if err := copyFile(src, dst, 0644); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Failed to read dst file: %v", err)
	}

	if string(got) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", string(got), string(content))
	}
}


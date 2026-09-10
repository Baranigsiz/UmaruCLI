package doctor

import (
	"testing"
)

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go version go1.24.0 windows/amd64", "1.24.0"},
		{"git version 2.44.0.windows.1", "2.44.0"},
		{"v20.10.0\n", "20.10.0"},
		{"Python 3.12.2", "3.12.2"},
		{"cargo 1.80.0 (376290e 2024-07-16)", "1.80.0"},
		{"Docker version 27.0.3, build 7d4eb23", "27.0.3"},
		{"10.8.2", "10.8.2"},
		{"", "unknown"},
		{"   \n\t  ", "unknown"},
	}

	for _, tt := range tests {
		result := CleanVersion(tt.input)
		if result != tt.expected {
			t.Errorf("CleanVersion(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAppendUnique(t *testing.T) {
	slice := []string{"node", "git"}
	slice = appendUnique(slice, "node")
	if len(slice) != 2 {
		t.Errorf("expected length 2 after duplicate append, got %d", len(slice))
	}

	slice = appendUnique(slice, "python")
	if len(slice) != 3 {
		t.Errorf("expected length 3 after unique append, got %d", len(slice))
	}
}

func TestCalculateReadiness_AllTools(t *testing.T) {
	mockTools := map[string]ToolCheck{
		"go":           {Status: StatusOk},
		"node.js":      {Status: StatusOk},
		"npm":          {Status: StatusOk},
		"bun":          {Status: StatusOk},
		"python":       {Status: StatusOk},
		"pip":          {Status: StatusOk},
		"cargo (rust)": {Status: StatusOk},
	}

	readiness := calculateReadiness(mockTools)
	if len(readiness) == 0 {
		t.Fatalf("expected template categories, got none")
	}

	for _, r := range readiness {
		if !r.IsReady {
			t.Errorf("category %s should be ready when all tools present, but got %d/%d (missing: %v)",
				r.Category, r.Ready, r.Total, r.Missing)
		}
	}
}

func TestCalculateReadiness_MissingBun(t *testing.T) {
	mockTools := map[string]ToolCheck{
		"go":           {Status: StatusOk},
		"node.js":      {Status: StatusOk},
		"npm":          {Status: StatusOk},
		"python":       {Status: StatusOk},
		"pip":          {Status: StatusOk},
		"cargo (rust)": {Status: StatusOk},
	}

	readiness := calculateReadiness(mockTools)
	for _, r := range readiness {
		if r.Category == "Node/TypeScript (Express, Fastify, Hono, NestJS, React, Vue, Svelte, Next, Astro)" {
			if r.IsReady {
				t.Errorf("expected Node/TypeScript not to be fully ready when bun is missing")
			}
			foundBun := false
			for _, m := range r.Missing {
				if m == "bun" {
					foundBun = true
					break
				}
			}
			if !foundBun {
				t.Errorf("expected 'bun' in missing tools list, got %v", r.Missing)
			}
		}
	}
}

func TestCalculateReadiness_MissingRust(t *testing.T) {
	mockTools := map[string]ToolCheck{
		"go":      {Status: StatusOk},
		"node.js": {Status: StatusOk},
		"npm":     {Status: StatusOk},
	}

	readiness := calculateReadiness(mockTools)
	var foundRust bool
	for _, r := range readiness {
		if r.Category == "Rust (Actix, Axum)" {
			foundRust = true
			if r.IsReady {
				t.Errorf("Rust templates should not be ready when cargo is missing")
			}
			if r.Ready != 0 {
				t.Errorf("expected 0 ready Rust templates, got %d", r.Ready)
			}
		}
	}

	if !foundRust {
		t.Errorf("expected Rust category in readiness list")
	}
}

func TestRunDiagnostics_Sanity(t *testing.T) {
	report := RunDiagnostics("v1.6.0-test")

	if report.System.OS == "" {
		t.Errorf("expected system OS to be populated")
	}
	if report.System.Arch == "" {
		t.Errorf("expected system Arch to be populated")
	}
	if len(report.Tools) == 0 {
		t.Errorf("expected non-empty tools list")
	}
	if len(report.Templates) == 0 {
		t.Errorf("expected non-empty templates readiness list")
	}
}

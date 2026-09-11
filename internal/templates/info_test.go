package templates

import (
	"strings"
	"testing"
)

func TestGetTemplatePorts(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"go-fiber", "3000"},
		{"fastify-api", "3000"},
		{"python-fastapi", "8000"},
		{"react-vite-ts", "5173"},
		{"fullstack-ts-monorepo", "8080"},
	}

	for _, tt := range tests {
		ports := GetTemplatePorts(tt.id)
		if len(ports) == 0 {
			t.Errorf("expected ports for %s, got none", tt.id)
		}
		found := false
		for _, p := range ports {
			if strings.Contains(p, tt.expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected port %s for %s, got %v", tt.expected, tt.id, ports)
		}
	}
}

func TestGetSupportedAddons(t *testing.T) {
	addons := GetSupportedAddons("go-fiber")
	if len(addons) == 0 {
		t.Fatalf("expected addons for go-fiber")
	}

	foundPG := false
	for _, a := range addons {
		if strings.Contains(a, "PostgreSQL") {
			foundPG = true
			break
		}
	}
	if !foundPG {
		t.Errorf("expected PostgreSQL in supported addons for go-fiber, got %v", addons)
	}
}

func TestGenerateTemplateTree(t *testing.T) {
	tree, count, err := GenerateTemplateTree("go-fiber")
	if err != nil {
		t.Fatalf("unexpected error generating tree: %v", err)
	}

	if count == 0 {
		t.Errorf("expected non-zero file count, got %d", count)
	}

	if !strings.HasPrefix(tree, "go-fiber/\n") {
		t.Errorf("tree should start with root directory, got: %s", tree)
	}

	// Should not show template.json
	if strings.Contains(tree, "template.json") {
		t.Errorf("tree should omit template.json")
	}

	// Should strip .tmpl
	if strings.Contains(tree, ".tmpl") {
		t.Errorf("tree should strip .tmpl suffixes for display")
	}

	// Should contain main.go
	if !strings.Contains(tree, "main.go") {
		t.Errorf("expected main.go in go-fiber tree, got:\n%s", tree)
	}
}

func TestGetTemplateInfo(t *testing.T) {
	info, err := GetTemplateInfo("fullstack-ts-monorepo")
	if err != nil {
		t.Fatalf("unexpected error getting template info: %v", err)
	}

	if info.Config.ID != "fullstack-ts-monorepo" {
		t.Errorf("expected ID fullstack-ts-monorepo, got %s", info.Config.ID)
	}

	if len(info.Ports) < 2 {
		t.Errorf("expected multi-port configuration for fullstack monorepo, got %v", info.Ports)
	}

	if !strings.Contains(info.FileTree, "apps/") {
		t.Errorf("expected apps/ directory in monorepo tree")
	}
}

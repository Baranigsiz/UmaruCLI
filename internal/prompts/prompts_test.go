package prompts

import (
	"testing"
	"umaru/internal/generator"
	"umaru/internal/templates"
)

func TestRun_NonInteractiveFullySpecified(t *testing.T) {
	t.Run("GoTemplateWithoutAddons", func(t *testing.T) {
		res, err := Run("my-go-app", "go-fiber", "", generator.AddonConfig{}, true)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if res.ProjectName != "my-go-app" {
			t.Errorf("Expected project name 'my-go-app', got '%s'", res.ProjectName)
		}
		if res.Template.ID != "go-fiber" {
			t.Errorf("Expected template ID 'go-fiber', got '%s'", res.Template.ID)
		}
	})

	t.Run("GoTemplateWithAddons", func(t *testing.T) {
		addons := generator.AddonConfig{
			Database: "postgres",
			Redis:    true,
		}
		res, err := Run("my-go-db", "go-fiber", "", addons, false)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if res.Addons.Database != "postgres" || !res.Addons.Redis {
			t.Errorf("Expected specified addons to be preserved, got %+v", res.Addons)
		}
	})

	t.Run("NodeTemplateWithExplicitPackageManager", func(t *testing.T) {
		res, err := Run("my-react-app", "react-vite-ts", "pnpm", generator.AddonConfig{}, true)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if res.PackageManager != "pnpm" {
			t.Errorf("Expected package manager 'pnpm', got '%s'", res.PackageManager)
		}
	})

	t.Run("BunTemplateAutoSelectsBun", func(t *testing.T) {
		res, err := Run("my-bun-app", "bun-elysia", "", generator.AddonConfig{}, true)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if res.PackageManager != "bun" {
			t.Errorf("Expected auto-selected package manager 'bun', got '%s'", res.PackageManager)
		}
	})

	t.Run("TemplateWithoutAddonSupportSatisfied", func(t *testing.T) {
		res, err := Run("my-desktop-app", "tauri-desktop", "npm", generator.AddonConfig{}, false)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if res.Template.ID != "tauri-desktop" {
			t.Errorf("Expected template tauri-desktop, got %s", res.Template.ID)
		}
	})

	t.Run("InvalidTemplateIDReturnsError", func(t *testing.T) {
		_, err := Run("my-bad-app", "invalid-template-xyz", "npm", generator.AddonConfig{}, true)
		if err == nil {
			t.Errorf("Expected error for invalid template ID, got nil")
		}
	})
}

func TestFilterTemplatesByKeyword(t *testing.T) {
	mockTemplates := []templates.TemplateConfig{
		{ID: "go-fiber", Name: "Go Fiber API", Description: "High performance web framework"},
		{ID: "react-vite-ts", Name: "React + Vite + TS", Description: "Modern frontend SPA"},
		{ID: "rust-axum", Name: "Rust Axum API", Description: "Ergonomic and modular web framework"},
		{ID: "tauri-desktop", Name: "Tauri v2 Desktop App", Description: "Cross-platform desktop with Rust"},
	}

	// 1. Empty query returns all
	if len(FilterTemplatesByKeyword(mockTemplates, "")) != 4 {
		t.Errorf("Expected all 4 templates for empty query")
	}

	// 2. Search by ID substring
	fiberRes := FilterTemplatesByKeyword(mockTemplates, "fiber")
	if len(fiberRes) != 1 || fiberRes[0].ID != "go-fiber" {
		t.Errorf("Expected go-fiber for 'fiber', got: %v", fiberRes)
	}

	// 3. Search by Name substring
	rustRes := FilterTemplatesByKeyword(mockTemplates, "Rust")
	if len(rustRes) != 2 {
		t.Errorf("Expected 2 rust templates for 'Rust', got: %d", len(rustRes))
	}

	// 4. Search by Description substring
	deskRes := FilterTemplatesByKeyword(mockTemplates, "desktop")
	if len(deskRes) != 1 || deskRes[0].ID != "tauri-desktop" {
		t.Errorf("Expected tauri-desktop for 'desktop', got: %v", deskRes)
	}

	// 5. Search non-matching
	noneRes := FilterTemplatesByKeyword(mockTemplates, "nonexistent999")
	if len(noneRes) != 0 {
		t.Errorf("Expected 0 results for nonexistent query, got: %d", len(noneRes))
	}
}

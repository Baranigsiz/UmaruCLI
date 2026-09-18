package prompts

import (
	"testing"
	"umaru/internal/generator"
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

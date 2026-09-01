package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddonConfig_HasAddons(t *testing.T) {
	tests := []struct {
		config   AddonConfig
		expected bool
	}{
		{AddonConfig{}, false},
		{AddonConfig{Database: "none", Auth: "none", Redis: false}, false},
		{AddonConfig{Database: "postgres"}, true},
		{AddonConfig{Auth: "jwt"}, true},
		{AddonConfig{Redis: true}, true},
		{AddonConfig{Database: "sqlite", Auth: "jwt", Redis: true}, true},
	}

	for _, tt := range tests {
		got := tt.config.HasAddons()
		if got != tt.expected {
			t.Errorf("HasAddons(%+v) = %v, want %v", tt.config, got, tt.expected)
		}
	}
}

func TestGetAddonFiles(t *testing.T) {
	cfg := ProjectConfig{
		TargetDir: "sample-dir",
		Template:  "go-fiber",
		Addons: AddonConfig{
			Database: "postgres",
			Auth:     "jwt",
			Redis:    true,
		},
	}

	files := GetAddonFiles(cfg)
	if len(files) != 3 {
		t.Fatalf("Expected 3 addon files, got %d: %v", len(files), files)
	}

	expected := []string{
		filepath.Join("sample-dir", "internal", "database", "postgres.go"),
		filepath.Join("sample-dir", "internal", "middleware", "auth.go"),
		filepath.Join("sample-dir", "internal", "cache", "redis.go"),
	}

	for i, f := range files {
		if f != expected[i] {
			t.Errorf("File[%d] = %s, want %s", i, f, expected[i])
		}
	}
}

func TestGenerateAddons_GoFiber(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "fiber-with-addons")

	config, err := ResolveProjectConfig(targetPath, "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}

	config.Addons = AddonConfig{
		Database: "postgres",
		Auth:     "jwt",
		Redis:    true,
	}

	// Run Generate (which internally calls GenerateAddons)
	err = Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generated addon files exist
	dbFile := filepath.Join(targetPath, "internal", "database", "postgres.go")
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", dbFile)
	}

	authFile := filepath.Join(targetPath, "internal", "middleware", "auth.go")
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", authFile)
	}

	redisFile := filepath.Join(targetPath, "internal", "cache", "redis.go")
	if _, err := os.Stat(redisFile); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", redisFile)
	}
}

package generator

import (
	"os"
	"path/filepath"
	"strings"
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
		{AddonConfig{Docker: true}, true},
		{AddonConfig{Database: "sqlite", Auth: "jwt", Redis: true, Docker: true}, true},
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

func TestGenerateAddons_PythonFastAPI(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "fastapi-with-addons")

	config, err := ResolveProjectConfig(targetPath, "python-fastapi")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}

	config.Addons = AddonConfig{
		Database: "postgres",
		Auth:     "jwt",
		Redis:    true,
	}

	err = Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify requirements.txt has been injected with required dependencies
	reqPath := filepath.Join(targetPath, "requirements.txt")
	content, err := os.ReadFile(reqPath)
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}
	contentStr := string(content)

	expectedDeps := []string{"asyncpg>=0.29.0", "python-jose[cryptography]>=3.3.0", "passlib[bcrypt]>=1.7.4", "redis>=5.0.0"}
	for _, dep := range expectedDeps {
		if !strings.Contains(contentStr, dep) {
			t.Errorf("Expected requirements.txt to contain %q, but it was missing", dep)
		}
	}
}

func TestGenerateAddons_NodeExpress(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "express-with-addons")

	config, err := ResolveProjectConfig(targetPath, "node-express")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}

	config.Addons = AddonConfig{
		Database: "postgres",
		Auth:     "jwt",
		Redis:    true,
	}

	err = Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify package.json has injected dependencies
	pkgPath := filepath.Join(targetPath, "package.json")
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	contentStr := string(content)

	expectedDeps := []string{`"pg": "^8.12.0"`, `"jsonwebtoken": "^9.0.2"`, `"ioredis": "^5.4.1"`, `"@types/pg": "^8.11.6"`, `"@types/jsonwebtoken": "^9.0.6"`, `"@types/ioredis": "^5.0.0"`}
	for _, dep := range expectedDeps {
		if !strings.Contains(contentStr, dep) {
			t.Errorf("Expected package.json to contain %q, but it was missing", dep)
		}
	}
}

func TestGenerateAddons_FrameworkSpecificAuth(t *testing.T) {
	tempDir := t.TempDir()

	// Test Hono
	honoPath := filepath.Join(tempDir, "hono-app")
	honoConfig, err := ResolveProjectConfig(honoPath, "hono-api")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	honoConfig.Addons = AddonConfig{Auth: "jwt"}
	if err := Generate(honoConfig); err != nil {
		t.Fatalf("Generate hono-api failed: %v", err)
	}
	honoAuthPath := filepath.Join(honoPath, "src", "middlewares", "auth.middleware.ts")
	honoAuth, err := os.ReadFile(honoAuthPath)
	if err != nil {
		t.Fatalf("Failed to read hono auth middleware: %v", err)
	}
	if !strings.Contains(string(honoAuth), "from 'hono'") {
		t.Errorf("Hono auth middleware should import from 'hono'")
	}
	if strings.Contains(string(honoAuth), "from 'express'") {
		t.Errorf("Hono auth middleware should NOT import from 'express'")
	}

	// Test Fastify
	fastifyPath := filepath.Join(tempDir, "fastify-app")
	fastifyConfig, err := ResolveProjectConfig(fastifyPath, "fastify-api")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	fastifyConfig.Addons = AddonConfig{Auth: "jwt"}
	if err := Generate(fastifyConfig); err != nil {
		t.Fatalf("Generate fastify-api failed: %v", err)
	}
	fastifyAuthPath := filepath.Join(fastifyPath, "src", "middlewares", "auth.middleware.ts")
	fastifyAuth, err := os.ReadFile(fastifyAuthPath)
	if err != nil {
		t.Fatalf("Failed to read fastify auth middleware: %v", err)
	}
	if !strings.Contains(string(fastifyAuth), "from 'fastify'") {
		t.Errorf("Fastify auth middleware should import from 'fastify'")
	}
	if strings.Contains(string(fastifyAuth), "from 'express'") {
		t.Errorf("Fastify auth middleware should NOT import from 'express'")
	}
}

func TestGenerateAddons_Docker(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		framework string
		needle    string
	}{
		{"go-fiber", "golang:1.24-alpine"},
		{"node-express", "node:20-alpine"},
		{"bun-elysia", "oven/bun:1"},
		{"python-fastapi", "python:3.11-slim"},
	}

	for _, tc := range testCases {
		t.Run(tc.framework, func(t *testing.T) {
			projPath := filepath.Join(tempDir, tc.framework+"-docker")
			cfg, err := ResolveProjectConfig(projPath, tc.framework)
			if err != nil {
				t.Fatalf("ResolveProjectConfig failed: %v", err)
			}
			cfg.Addons = AddonConfig{Docker: true}

			if err := Generate(cfg); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			// Verify Dockerfile
			dockerfilePath := filepath.Join(projPath, "Dockerfile")
			dfContent, err := os.ReadFile(dockerfilePath)
			if err != nil {
				t.Fatalf("Dockerfile not found: %v", err)
			}
			if !strings.Contains(string(dfContent), tc.needle) {
				t.Errorf("Dockerfile for %s should contain '%s'", tc.framework, tc.needle)
			}

			// Verify docker-compose.yml
			composePath := filepath.Join(projPath, "docker-compose.yml")
			if _, err := os.Stat(composePath); err != nil {
				t.Errorf("docker-compose.yml not found: %v", err)
			}

			// Verify .dockerignore
			ignorePath := filepath.Join(projPath, ".dockerignore")
			if _, err := os.Stat(ignorePath); err != nil {
				t.Errorf(".dockerignore not found: %v", err)
			}
		})
	}
}


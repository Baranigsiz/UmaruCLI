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
		{AddonConfig{CI: true}, true},
		{AddonConfig{Database: "sqlite", Auth: "jwt", Redis: true, Docker: true, CI: true}, true},
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
	if len(files) != 5 {
		t.Fatalf("Expected 5 addon files, got %d: %v", len(files), files)
	}

	expected := []string{
		filepath.Join("sample-dir", "internal", "database", "postgres.go"),
		filepath.Join("sample-dir", "internal", "middleware", "auth.go"),
		filepath.Join("sample-dir", "internal", "cache", "redis.go"),
		filepath.Join("sample-dir", ".env.example"),
		filepath.Join("sample-dir", ".env"),
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

func TestGenerateAddons_GoSQLite(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "fiber-with-sqlite")

	config, err := ResolveProjectConfig(targetPath, "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	config.Addons = AddonConfig{Database: "sqlite"}

	err = Generate(config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	dbFile := filepath.Join(targetPath, "internal", "database", "sqlite.go")
	data, err := os.ReadFile(dbFile)
	if err != nil {
		t.Fatalf("Expected %s to exist: %v", dbFile, err)
	}

	content := string(data)
	if !strings.Contains(content, "modernc.org/sqlite") {
		t.Errorf("Expected sqlite.go to use pure-Go modernc.org/sqlite, got:\n%s", content)
	}
	if strings.Contains(content, "mattn/go-sqlite3") {
		t.Errorf("sqlite.go should not contain CGO dependency mattn/go-sqlite3")
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
		{"rust-axum", "rust:1.80-slim-bullseye"},
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

func TestGenerateAddons_CI(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		framework string
		needle    string
	}{
		{"go-fiber", "actions/setup-go@v5"},
		{"node-express", "actions/setup-node@v4"},
		{"bun-elysia", "oven-sh/setup-bun@v2"},
		{"python-fastapi", "actions/setup-python@v5"},
		{"rust-axum", "dtolnay/rust-toolchain@stable"},
	}

	for _, tc := range testCases {
		t.Run(tc.framework, func(t *testing.T) {
			projPath := filepath.Join(tempDir, tc.framework+"-ci")
			cfg, err := ResolveProjectConfig(projPath, tc.framework)
			if err != nil {
				t.Fatalf("ResolveProjectConfig failed: %v", err)
			}
			cfg.Addons = AddonConfig{CI: true}

			if err := Generate(cfg); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			ciWorkflowPath := filepath.Join(projPath, ".github", "workflows", "ci.yml")
			ciContent, err := os.ReadFile(ciWorkflowPath)
			if err != nil {
				t.Fatalf("ci.yml not found: %v", err)
			}
			if !strings.Contains(string(ciContent), tc.needle) {
				t.Errorf("ci.yml for %s should contain '%s'", tc.framework, tc.needle)
			}
		})
	}
}

func TestGenerateAddons_DockerWithServices(t *testing.T) {
	tempDir := t.TempDir()
	projPath := filepath.Join(tempDir, "fiber-full-docker")

	cfg, err := ResolveProjectConfig(projPath, "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	cfg.Addons = AddonConfig{
		Docker:   true,
		Database: "postgres",
		Redis:    true,
	}

	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	composePath := filepath.Join(projPath, "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("docker-compose.yml not found: %v", err)
	}

	composeStr := string(content)
	needles := []string{
		"postgres:16-alpine",
		"redis:7-alpine",
		"5432:5432",
		"6379:6379",
		"depends_on:",
		"postgres",
		"redis",
		"volumes:",
		"_pgdata",
		"_redisdata",
	}

	for _, n := range needles {
		if !strings.Contains(composeStr, n) {
			t.Errorf("docker-compose.yml should contain '%s'", n)
		}
	}
}

func TestGenerateAddons_EnvVariables(t *testing.T) {
	tempDir := t.TempDir()
	projPath := filepath.Join(tempDir, "fiber-env-test")

	cfg, err := ResolveProjectConfig(projPath, "go-fiber")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	cfg.Addons = AddonConfig{
		Database: "postgres",
		Auth:     "jwt",
		Redis:    true,
	}

	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify .env.example
	envExamplePath := filepath.Join(projPath, ".env.example")
	dataExample, err := os.ReadFile(envExamplePath)
	if err != nil {
		t.Fatalf(".env.example not found: %v", err)
	}

	exampleStr := string(dataExample)
	expectedVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "JWT_SECRET", "REDIS_ADDR"}
	for _, v := range expectedVars {
		if !strings.Contains(exampleStr, v) {
			t.Errorf(".env.example should contain '%s'", v)
		}
	}

	// Verify .env was automatically created as well
	envPath := filepath.Join(projPath, ".env")
	dataEnv, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf(".env was not automatically created: %v", err)
	}

	envStr := string(dataEnv)
	for _, v := range expectedVars {
		if !strings.Contains(envStr, v) {
			t.Errorf(".env should contain '%s'", v)
		}
	}
}

func TestAppendDockerCompose_ExistingServiceVolumes(t *testing.T) {
	tempDir := t.TempDir()
	composePath := filepath.Join(tempDir, "docker-compose.yml")

	initialYAML := `version: '3.8'

services:
  api:
    build: .
    ports:
      - "8000:8000"
    restart: unless-stopped

  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8001:8000"
    volumes:
      - chroma_data:/chroma/chroma
    restart: unless-stopped

volumes:
  chroma_data:
`
	if err := os.WriteFile(composePath, []byte(initialYAML), 0644); err != nil {
		t.Fatalf("failed to write initial compose: %v", err)
	}

	cfg := ProjectConfig{
		TargetDir: tempDir,
		SafeName:  "test-app",
		Addons: AddonConfig{
			Database: "postgres",
			Redis:    true,
		},
	}

	if err := appendDockerComposeServices(tempDir, cfg); err != nil {
		t.Fatalf("appendDockerComposeServices failed: %v", err)
	}

	updated, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("failed to read updated compose: %v", err)
	}
	updatedStr := string(updated)

	// Verify chromadb service volume was NOT mangled
	if !strings.Contains(updatedStr, "- chroma_data:/chroma/chroma") {
		t.Errorf("chromadb service volumes mount was corrupted/lost:\n%s", updatedStr)
	}

	// Verify top-level volumes contains chroma_data AND test-app_pgdata AND test-app_redisdata
	if !strings.Contains(updatedStr, "chroma_data:") {
		t.Errorf("top-level chroma_data was lost:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "test-app_pgdata:") {
		t.Errorf("test-app_pgdata was not added to volumes:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "test-app_redisdata:") {
		t.Errorf("test-app_redisdata was not added to volumes:\n%s", updatedStr)
	}

	// Verify services were added
	if !strings.Contains(updatedStr, "postgres:") || !strings.Contains(updatedStr, "redis:") {
		t.Errorf("postgres or redis service was not added:\n%s", updatedStr)
	}
}

func TestGenerateAddons_MonorepoCILocation(t *testing.T) {
	tempDir := t.TempDir()
	projPath := filepath.Join(tempDir, "monorepo-ci-test")

	cfg, err := ResolveProjectConfig(projPath, "fullstack-go-react")
	if err != nil {
		t.Fatalf("ResolveProjectConfig failed: %v", err)
	}
	cfg.Addons = AddonConfig{
		CI:       true,
		Database: "postgres",
	}

	if err := Generate(cfg); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// CI workflow MUST be at repository root
	rootCI := filepath.Join(projPath, ".github", "workflows", "ci.yml")
	if _, err := os.Stat(rootCI); os.IsNotExist(err) {
		t.Errorf("Expected root CI workflow %s to exist", rootCI)
	}

	// CI workflow MUST NOT be nested in apps/api
	wrongCI := filepath.Join(projPath, "apps", "api", ".github", "workflows", "ci.yml")
	if _, err := os.Stat(wrongCI); err == nil {
		t.Errorf("CI workflow was incorrectly placed in apps/api: %s", wrongCI)
	}

	// Root docker-compose.yml should be updated with postgres
	rootCompose := filepath.Join(projPath, "docker-compose.yml")
	content, err := os.ReadFile(rootCompose)
	if err != nil {
		t.Fatalf("failed to read root compose: %v", err)
	}
	if !strings.Contains(string(content), "postgres:") {
		t.Errorf("root docker-compose.yml was not updated with postgres service:\n%s", string(content))
	}
}



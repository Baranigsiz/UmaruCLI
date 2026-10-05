package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_FullDeveloperLifecycle executes a comprehensive, end-to-end integration test
// that mimics the entire real-world lifecycle of an Umaru CLI project:
// 1. Scaffolding a project with `umaru init`
// 2. Injecting modular addons with `umaru add`
// 3. Generating CRUD clean architecture resources with `umaru generate`
// 4. Running health & security diagnostics with `umaru audit`
// 5. Resolving test runner commands with `umaru test`
// 6. Generating production cloud deployment manifests with `umaru deploy`
// 7. Scanning and sanitizing workspace with `umaru clean`
func TestE2E_FullDeveloperLifecycle(t *testing.T) {
	tempWorkspace := t.TempDir()
	projectDir := filepath.Join(tempWorkspace, "shop-api")

	// ==========================================
	// STEP 1: umaru init (Scaffold Go Fiber API)
	// ==========================================
	t.Log("Step 1: Scaffolding project via 'umaru init'...")
	initOut, err := executeCommand("init", projectDir, "-t", "go-fiber", "--no-git", "--yes")
	if err != nil {
		t.Fatalf("Step 1 failed: 'umaru init' error: %v\nOutput: %s", err, initOut)
	}

	goModPath := filepath.Join(projectDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatalf("Step 1 assertion failed: go.mod not found at %s", goModPath)
	}

	goModBytes, _ := os.ReadFile(goModPath)
	if !strings.Contains(string(goModBytes), "module shop-api") {
		t.Errorf("Step 1 assertion failed: expected 'module shop-api' in go.mod, got:\n%s", string(goModBytes))
	}

	// ==========================================
	// STEP 2: umaru add (Inject Postgres, JWT, Redis)
	// ==========================================
	t.Log("Step 2: Injecting modular infrastructure addons via 'umaru add'...")
	addOut, err := executeCommand("add", "postgres", "jwt", "redis", "--dir", projectDir, "--force")
	if err != nil {
		t.Fatalf("Step 2 failed: 'umaru add' error: %v\nOutput: %s", err, addOut)
	}

	expectedAddonFiles := []string{
		filepath.Join(projectDir, "internal", "database", "postgres.go"),
		filepath.Join(projectDir, "internal", "middleware", "auth.go"),
		filepath.Join(projectDir, "internal", "cache", "redis.go"),
		filepath.Join(projectDir, ".env"),
		filepath.Join(projectDir, ".env.example"),
	}

	for _, ef := range expectedAddonFiles {
		if _, err := os.Stat(ef); os.IsNotExist(err) {
			t.Errorf("Step 2 assertion failed: expected addon file %s to exist", ef)
		}
	}

	// Verify JWT Secret in .env is not hardcoded placeholder
	envContent, _ := os.ReadFile(filepath.Join(projectDir, ".env"))
	if strings.Contains(string(envContent), "super-secret-key-change-in-production") {
		t.Errorf("Step 2 assertion failed: .env still contains insecure static placeholder JWT secret!")
	}
	if !strings.Contains(string(envContent), "JWT_SECRET=") {
		t.Errorf("Step 2 assertion failed: .env missing JWT_SECRET key")
	}

	// ==========================================
	// STEP 3: umaru generate resource Customer
	// ==========================================
	t.Log("Step 3: Generating clean architecture resource via 'umaru generate resource'...")
	genOut, err := executeCommand("generate", "resource", "Customer", "-d", projectDir, "--force")
	if err != nil {
		t.Fatalf("Step 3 failed: 'umaru generate' error: %v\nOutput: %s", err, genOut)
	}

	expectedResourceFiles := []string{
		filepath.Join(projectDir, "internal", "models", "customer.go"),
		filepath.Join(projectDir, "internal", "repository", "customer_repository.go"),
		filepath.Join(projectDir, "internal", "service", "customer_service.go"),
		filepath.Join(projectDir, "internal", "handlers", "customer.go"),
	}

	for _, rf := range expectedResourceFiles {
		if _, err := os.Stat(rf); os.IsNotExist(err) {
			t.Errorf("Step 3 assertion failed: expected resource file %s to exist", rf)
		}
	}

	// Verify that the generated repository imports the correct dynamic module "shop-api/internal/models"
	repoContent, _ := os.ReadFile(filepath.Join(projectDir, "internal", "repository", "customer_repository.go"))
	if !strings.Contains(string(repoContent), "\"shop-api/internal/models\"") {
		t.Errorf("Step 3 assertion failed: generated repository has incorrect import paths. Expected 'shop-api/internal/models', got:\n%s", string(repoContent))
	}

	// ==========================================
	// STEP 4: umaru audit (Diagnostics & Health)
	// ==========================================
	t.Log("Step 4: Running project health diagnostics via 'umaru audit'...")
	// Create mock git repo with .gitignore to pass audit security check
	_ = os.MkdirAll(filepath.Join(projectDir, ".git"), 0755)
	_ = os.WriteFile(filepath.Join(projectDir, ".gitignore"), []byte(".env\nbin/\n"), 0644)
	_ = os.WriteFile(filepath.Join(projectDir, "go.sum"), []byte(""), 0644)

	auditOut, err := executeCommand("audit", projectDir, "--no-network", "--json")
	if err != nil {
		t.Fatalf("Step 4 failed: 'umaru audit' error: %v\nOutput: %s", err, auditOut)
	}

	var auditReport struct {
		HealthScore int `json:"health_score"`
		Checks      []struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(auditOut), &auditReport); err != nil {
		t.Fatalf("Step 4 failed parsing audit JSON: %v, raw:\n%s", err, auditOut)
	}

	if auditReport.HealthScore < 80 {
		t.Errorf("Step 4 assertion failed: expected health score >= 80, got %d", auditReport.HealthScore)
	}
	for _, check := range auditReport.Checks {
		if check.Status == "fail" {
			t.Errorf("Step 4 assertion failed: unexpected audit failure: %s", check.Message)
		}
	}

	// ==========================================
	// STEP 5: umaru test (Universal Test Runner)
	// ==========================================
	t.Log("Step 5: Testing project discovery via 'umaru test --dry-run'...")
	testOut, err := executeCommand("test", "--dry-run", "--json", projectDir)
	if err != nil {
		t.Fatalf("Step 5 failed: 'umaru test' error: %v\nOutput: %s", err, testOut)
	}

	var testConfig struct {
		Language    string   `json:"language"`
		Command     []string `json:"command"`
		CommandLine string   `json:"command_line"`
	}
	if err := json.Unmarshal([]byte(testOut), &testConfig); err != nil {
		t.Fatalf("Step 5 failed parsing test JSON: %v, raw:\n%s", err, testOut)
	}

	if testConfig.Language != "go" {
		t.Errorf("Step 5 assertion failed: expected language 'go', got '%s'", testConfig.Language)
	}
	if !strings.Contains(testConfig.CommandLine, "go test") {
		t.Errorf("Step 5 assertion failed: expected 'go test' in command, got '%s'", testConfig.CommandLine)
	}

	// ==========================================
	// STEP 6: umaru deploy (Railway & Fly Cloud)
	// ==========================================
	t.Log("Step 6: Generating cloud deployment manifests via 'umaru deploy'...")
	deployOut, err := executeCommand("deploy", "railway", "-d", projectDir, "--json")
	if err != nil {
		t.Fatalf("Step 6 failed: 'umaru deploy railway' error: %v\nOutput: %s", err, deployOut)
	}

	railwayPath := filepath.Join(projectDir, "railway.json")
	if _, err := os.Stat(railwayPath); os.IsNotExist(err) {
		t.Fatalf("Step 6 assertion failed: railway.json was not written")
	}

	railwayBytes, _ := os.ReadFile(railwayPath)
	if !strings.Contains(string(railwayBytes), "buildCommand") {
		t.Errorf("Step 6 assertion failed: railway.json missing buildCommand: %s", string(railwayBytes))
	}
	if !strings.Contains(string(railwayBytes), "\"startCommand\": \"./server\"") {
		t.Errorf("Step 6 assertion failed: railway.json missing startCommand ./server: %s", string(railwayBytes))
	}

	// ==========================================
	// STEP 7: umaru clean (Disk Sanitizer)
	// ==========================================
	t.Log("Step 7: Testing workspace cleaning via 'umaru clean --dry-run'...")
	// Create mock removable directory (e.g. dist)
	distDir := filepath.Join(projectDir, "dist")
	_ = os.MkdirAll(distDir, 0755)
	_ = os.WriteFile(filepath.Join(distDir, "bundle.js"), []byte("console.log(1)"), 0644)

	cleanOut, err := executeCommand("clean", projectDir, "--dry-run", "--json")
	if err != nil {
		t.Fatalf("Step 7 failed: 'umaru clean' error: %v\nOutput: %s", err, cleanOut)
	}

	var cleanReport struct {
		TotalSizeBytes int64 `json:"total_size_bytes"`
		Items          []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(cleanOut), &cleanReport); err != nil {
		t.Fatalf("Step 7 failed parsing clean JSON: %v, raw:\n%s", err, cleanOut)
	}

	foundDist := false
	for _, item := range cleanReport.Items {
		if item.Name == "dist" {
			foundDist = true
			break
		}
	}
	if !foundDist {
		t.Errorf("Step 7 assertion failed: expected 'dist' to be detected in clean scan")
	}

	// Verify project source files are completely untouched
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("Step 7 CRITICAL: go.mod was deleted during clean check!")
	}
	t.Log("✅ E2E Life Cycle Test completed successfully: All 7 lifecycle phases passed!")
}

// TestE2E_NodeDrizzleLifecycle verifies the end-to-end developer flow for Node/TypeScript:
// init (fastify-api) -> add drizzle -> generate resource -> test discovery -> deploy
func TestE2E_NodeDrizzleLifecycle(t *testing.T) {
	tempWorkspace := t.TempDir()
	projectDir := filepath.Join(tempWorkspace, "ts-service")

	// 1. Init Fastify API with pnpm (using --skip-install for test environment)
	t.Log("Step 1: Scaffolding fastify-api via 'umaru init'...")
	initOut, err := executeCommand("init", projectDir, "-t", "fastify-api", "--pm", "pnpm", "--no-git", "--skip-install", "--yes")
	if err != nil {
		t.Fatalf("Step 1 failed: %v\nOutput: %s", err, initOut)
	}

	pkgPath := filepath.Join(projectDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		t.Fatalf("Step 1 assertion failed: package.json missing")
	}

	// 2. Add Drizzle ORM
	t.Log("Step 2: Injecting Drizzle ORM via 'umaru add drizzle'...")
	addOut, err := executeCommand("add", "drizzle", "--dir", projectDir, "--force")
	if err != nil {
		t.Fatalf("Step 2 failed: %v\nOutput: %s", err, addOut)
	}

	drizzleConfigFile := filepath.Join(projectDir, "drizzle.config.ts")
	if _, err := os.Stat(drizzleConfigFile); os.IsNotExist(err) {
		t.Errorf("Step 2 assertion failed: drizzle.config.ts was not generated")
	}

	// 3. Generate resource
	t.Log("Step 3: Generating clean architecture resource via 'umaru generate resource Article'...")
	genOut, err := executeCommand("generate", "resource", "Article", "-d", projectDir, "--force")
	if err != nil {
		t.Fatalf("Step 3 failed: %v\nOutput: %s", err, genOut)
	}

	expectedTSFiles := []string{
		filepath.Join(projectDir, "src", "models", "article.model.ts"),
		filepath.Join(projectDir, "src", "services", "article.service.ts"),
		filepath.Join(projectDir, "src", "controllers", "article.controller.ts"),
		filepath.Join(projectDir, "src", "routes", "article.routes.ts"),
	}
	for _, f := range expectedTSFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Step 3 assertion failed: expected %s to exist", f)
		}
	}

	// 4. Test discovery via umaru test
	t.Log("Step 4: Testing project test command detection via 'umaru test --dry-run'...")
	testOut, err := executeCommand("test", "--dry-run", "--json", projectDir)
	if err != nil {
		t.Fatalf("Step 4 failed: %v\nOutput: %s", err, testOut)
	}

	var testConfig struct {
		Language    string `json:"language"`
		CommandLine string `json:"command_line"`
	}
	_ = json.Unmarshal([]byte(testOut), &testConfig)
	if testConfig.Language != "node" {
		t.Errorf("Step 4 assertion failed: expected language 'node', got '%s'", testConfig.Language)
	}

	// 5. Deploy Render Blueprint
	t.Log("Step 5: Generating Render Blueprint deployment via 'umaru deploy render'...")
	deployOut, err := executeCommand("deploy", "render", "-d", projectDir, "--json")
	if err != nil {
		t.Fatalf("Step 5 failed: %v\nOutput: %s", err, deployOut)
	}

	renderYaml := filepath.Join(projectDir, "render.yaml")
	if _, err := os.Stat(renderYaml); os.IsNotExist(err) {
		t.Errorf("Step 5 assertion failed: render.yaml missing")
	}

	t.Log("✅ Node/Drizzle E2E Life Cycle Test completed successfully!")
}


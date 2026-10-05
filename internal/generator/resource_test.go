package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResourceHelpers(t *testing.T) {
	tests := []struct {
		input  string
		pascal string
		camel  string
		plural string
	}{
		{"user", "User", "user", "users"},
		{"user_profile", "UserProfile", "userProfile", "user-profiles"},
		{"product-category", "ProductCategory", "productCategory", "product-categories"},
		{"order_item", "OrderItem", "orderItem", "order-items"},
		{"status", "Status", "status", "statuses"},
	}

	for _, tt := range tests {
		gotPascal := ToPascalCase(tt.input)
		if gotPascal != tt.pascal {
			t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, gotPascal, tt.pascal)
		}

		gotCamel := ToCamelCase(tt.input)
		if gotCamel != tt.camel {
			t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, gotCamel, tt.camel)
		}
	}
}

func TestGenerateResource_GoFiber(t *testing.T) {
	tempDir := t.TempDir()

	// Mock Go Fiber project
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24\nrequire github.com/gofiber/fiber/v2 v2.52.0"), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, "cmd", "api"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "cmd", "api", "main.go"), []byte("package main"), 0644)

	res, err := GenerateResource(ResourceConfig{
		Name:      "Product",
		TargetDir: tempDir,
	})
	if err != nil {
		t.Fatalf("GenerateResource failed: %v", err)
	}

	if res.ResourceName != "Product" {
		t.Errorf("Expected ResourceName 'Product', got '%s'", res.ResourceName)
	}
	if len(res.Files) != 4 {
		t.Fatalf("Expected 4 files, got %d", len(res.Files))
	}

	expectedFiles := []string{
		filepath.Join("internal", "models", "product.go"),
		filepath.Join("internal", "repository", "product_repository.go"),
		filepath.Join("internal", "service", "product_service.go"),
		filepath.Join("internal", "handlers", "product.go"),
	}

	for _, ef := range expectedFiles {
		fullPath := filepath.Join(tempDir, ef)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", ef)
		}
	}

	// Verify model content contains Product struct
	modelBytes, _ := os.ReadFile(filepath.Join(tempDir, "internal", "models", "product.go"))
	if !strings.Contains(string(modelBytes), "type Product struct") {
		t.Errorf("Expected Product struct in model file")
	}

	// Verify repository uses dynamic module name from go.mod ("testapp")
	repoBytes, _ := os.ReadFile(filepath.Join(tempDir, "internal", "repository", "product_repository.go"))
	if !strings.Contains(string(repoBytes), "\"testapp/internal/models\"") {
		t.Errorf("Expected dynamic module import 'testapp/internal/models', got:\n%s", string(repoBytes))
	}

	// Test skip on existing file without force
	res2, err := GenerateResource(ResourceConfig{
		Name:      "Product",
		TargetDir: tempDir,
		Force:     false,
	})
	if err != nil {
		t.Fatalf("GenerateResource second run failed: %v", err)
	}
	for _, f := range res2.Files {
		if !strings.Contains(f.Action, "skipped") {
			t.Errorf("Expected action to be skipped, got %s for %s", f.Action, f.RelPath)
		}
	}
}

func TestGenerateResource_NodeExpress(t *testing.T) {
	tempDir := t.TempDir()

	// Mock Node project
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name":"test-app","dependencies":{"express":"^4.19.0"}}`), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, "src"), 0755)

	res, err := GenerateResource(ResourceConfig{
		Name:      "Order",
		TargetDir: tempDir,
	})
	if err != nil {
		t.Fatalf("GenerateResource failed: %v", err)
	}

	if len(res.Files) != 4 {
		t.Fatalf("Expected 4 files, got %d", len(res.Files))
	}

	orderModel := filepath.Join(tempDir, "src", "models", "order.model.ts")
	if _, err := os.Stat(orderModel); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", orderModel)
	}
}

func TestGenerateResource_PythonFastAPI(t *testing.T) {
	tempDir := t.TempDir()

	// Mock FastAPI project
	_ = os.MkdirAll(filepath.Join(tempDir, "app", "api"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte("fastapi\nuvicorn"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "app", "main.py"), []byte("from fastapi import FastAPI"), 0644)

	res, err := GenerateResource(ResourceConfig{
		Name:      "customer",
		TargetDir: tempDir,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("GenerateResource dry-run failed: %v", err)
	}

	if len(res.Files) != 2 {
		t.Fatalf("Expected 2 files for Python, got %d", len(res.Files))
	}

	// Dry run shouldn't write to disk
	schemaPath := filepath.Join(tempDir, "app", "schemas", "customer.py")
	if _, err := os.Stat(schemaPath); !os.IsNotExist(err) {
		t.Errorf("Dry run should not create file on disk: %s", schemaPath)
	}
}

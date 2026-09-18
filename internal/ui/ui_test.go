package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"umaru/internal/generator"
	"umaru/internal/templates"
)

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	_ = w.Close()
	os.Stdout = oldStdout
	out := <-outC
	_ = r.Close()
	return out
}

func TestPrintBanner(t *testing.T) {
	out := captureOutput(func() {
		PrintBanner()
	})

	if !strings.Contains(out, "UMARU CLI") {
		t.Errorf("PrintBanner() expected to contain 'UMARU CLI', got: %s", out)
	}
	if !strings.Contains(out, "Production-Ready Project Scaffolder") {
		t.Errorf("PrintBanner() expected to contain 'Production-Ready Project Scaffolder', got: %s", out)
	}
}

func TestPrintDryRunCard_WithoutAddons(t *testing.T) {
	cfg := generator.ProjectConfig{
		TargetDir:   "my-test-app",
		SafeName:    "my-test-app",
		ProjectName: "My Test App",
	}
	files := []string{"package.json", "src/index.ts", "README.md"}

	out := captureOutput(func() {
		PrintDryRunCard(cfg, "Node Express", files)
	})

	if !strings.Contains(out, "Dry-Run Mode") {
		t.Errorf("Expected Dry-Run Mode title, got: %s", out)
	}
	if !strings.Contains(out, "my-test-app") {
		t.Errorf("Expected target directory in output, got: %s", out)
	}
	if !strings.Contains(out, "Node Express") {
		t.Errorf("Expected template name in output, got: %s", out)
	}
	if !strings.Contains(out, "src/index.ts") {
		t.Errorf("Expected files listed, got: %s", out)
	}
}

func TestPrintDryRunCard_WithAddons(t *testing.T) {
	cfg := generator.ProjectConfig{
		TargetDir:   "my-full-app",
		SafeName:    "my-full-app",
		ProjectName: "My Full App",
		Addons: generator.AddonConfig{
			Database: "postgres",
			Auth:     "jwt",
			Redis:    true,
		},
	}
	files := []string{"main.go", "go.mod"}

	out := captureOutput(func() {
		PrintDryRunCard(cfg, "Go Fiber", files)
	})

	if !strings.Contains(out, "DB: postgres") {
		t.Errorf("Expected DB addon in dry run, got: %s", out)
	}
	if !strings.Contains(out, "Auth: jwt") {
		t.Errorf("Expected Auth addon in dry run, got: %s", out)
	}
	if !strings.Contains(out, "Cache: Redis") {
		t.Errorf("Expected Redis addon in dry run, got: %s", out)
	}
}

func TestPrintSuccessCard_VariousOptions(t *testing.T) {
	t.Run("StandardSuccess", func(t *testing.T) {
		cfg := generator.ProjectConfig{
			TargetDir:   "awesome-app",
			SafeName:    "awesome-app",
			ProjectName: "Awesome App",
			Addons: generator.AddonConfig{
				Database: "sqlite",
				Auth:     "jwt",
				Redis:    true,
				Docker:   true,
				CI:       true,
			},
		}

		out := captureOutput(func() {
			PrintSuccessCard(cfg, "React Vite TS", "npm run dev", []string{"npm", "install"}, true)
		})

		if !strings.Contains(out, "Project Scaffolding Complete!") {
			t.Errorf("Expected completion title, got: %s", out)
		}
		if !strings.Contains(out, "cd awesome-app") {
			t.Errorf("Expected cd step, got: %s", out)
		}
		if !strings.Contains(out, "npm install") {
			t.Errorf("Expected install step when skipInstall=true, got: %s", out)
		}
		if !strings.Contains(out, "npm run dev") {
			t.Errorf("Expected runCommand in next steps, got: %s", out)
		}
		if !strings.Contains(out, "Docker: Containerized") {
			t.Errorf("Expected Docker addon in success card, got: %s", out)
		}
		if !strings.Contains(out, "CI: GitHub Actions") {
			t.Errorf("Expected CI addon in success card, got: %s", out)
		}
	})

	t.Run("CurrentDirectoryTarget", func(t *testing.T) {
		cfg := generator.ProjectConfig{
			TargetDir:   ".",
			SafeName:    "current-dir-app",
			ProjectName: "Current Dir App",
		}

		out := captureOutput(func() {
			PrintSuccessCard(cfg, "Go CLI", "go run .", []string{"go", "mod", "tidy"}, false)
		})

		if strings.Contains(out, "cd .") {
			t.Errorf("Target '.' should not show 'cd .' step, got: %s", out)
		}
		if !strings.Contains(out, "1. go run .") {
			t.Errorf("Expected step 1 to be run command when target is ., got: %s", out)
		}
	})

	t.Run("DirectoryWithSpaces", func(t *testing.T) {
		cfg := generator.ProjectConfig{
			TargetDir:   "My Spaced Project",
			SafeName:    "my-spaced-project",
			ProjectName: "My Spaced Project",
		}

		out := captureOutput(func() {
			PrintSuccessCard(cfg, "Go Fiber", "go run .", nil, false)
		})

		if !strings.Contains(out, `cd "My Spaced Project"`) {
			t.Errorf("Expected quoted cd directory for paths with spaces, got: %s", out)
		}
	})
}

func TestPrintTemplateInfoCard(t *testing.T) {
	info := templates.TemplateInfo{
		Config: templates.TemplateConfig{
			ID:             "test-template",
			Name:           "Test Template",
			Description:    "A test template for UI tests",
			InstallCommand: []string{"npm", "install"},
			RunCommand:     "npm run dev",
		},
		Ports:            []string{"3000 (HTTP API)", "8080 (Admin)"},
		SupportedAddons:  []string{"PostgreSQL", "SQLite", "JWT", "Redis", "Docker", "CI/CD"},
		FileTree:         "test-template/\n├── src/\n│   └── index.ts\n└── package.json",
		TotalFiles:       3,
	}

	out := captureOutput(func() {
		PrintTemplateInfoCard(&info)
	})

	if !strings.Contains(out, "Test Template") {
		t.Errorf("Expected template name in info card, got: %s", out)
	}
	if !strings.Contains(out, "test-template") {
		t.Errorf("Expected template ID in info card, got: %s", out)
	}
	if !strings.Contains(out, "3000 (HTTP API)") {
		t.Errorf("Expected ports in info card, got: %s", out)
	}
	if !strings.Contains(out, "npm install") {
		t.Errorf("Expected install command in info card, got: %s", out)
	}
	if !strings.Contains(out, "Architecture Tree (3 files)") {
		t.Errorf("Expected architecture tree header, got: %s", out)
	}
}

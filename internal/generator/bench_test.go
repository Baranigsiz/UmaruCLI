package generator

import (
	"os"
	"path/filepath"
	"testing"
	"umaru/internal/templates"
)

func BenchmarkSlugify(b *testing.B) {
	input := "Türkiye Süper Lig ve Şampiyonlar Ligi 2026 - Modern Proje!"
	for b.Loop() {
		_ = Slugify(input)
	}
}

func BenchmarkAuditProjectAddons(b *testing.B) {
	tempDir := b.TempDir()
	// Create minimal Go project with a Dockerfile
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testbench\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "Dockerfile"), []byte("FROM alpine\n"), 0644)

	for b.Loop() {
		_, _ = AuditProjectAddons(tempDir)
	}
}

func BenchmarkTemplateFind(b *testing.B) {
	for b.Loop() {
		_, _ = templates.FindTemplateByID("go-fiber")
	}
}

func BenchmarkGenerateProject_GoFiber(b *testing.B) {
	cfg := ProjectConfig{
		ProjectName: "bench-fiber",
		SafeName:    "bench-fiber",
		ModuleName:  "bench-fiber",
		Template:    "go-fiber",
		License:     "MIT",
	}

	for b.Loop() {
		b.StopTimer()
		target := filepath.Join(b.TempDir(), "target")
		cfg.TargetDir = target
		b.StartTimer()

		_ = Generate(cfg)
	}
}

func BenchmarkGenerateProject_ReactVite(b *testing.B) {
	cfg := ProjectConfig{
		ProjectName: "bench-react",
		SafeName:    "bench-react",
		ModuleName:  "bench-react",
		Template:    "react-vite-ts",
		License:     "MIT",
	}

	for b.Loop() {
		b.StopTimer()
		target := filepath.Join(b.TempDir(), "target")
		cfg.TargetDir = target
		b.StartTimer()

		_ = Generate(cfg)
	}
}


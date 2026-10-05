package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckEnvironmentVariables_Parity(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, ".env.example"), []byte("PORT=8080\nDB_URL=postgres://...\nJWT_SECRET=abc"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("PORT=8080\nDB_URL=postgres://...\nJWT_SECRET=abc"), 0644)

	checks := checkEnvironmentVariables(tempDir)
	if len(checks) == 0 {
		t.Fatalf("expected checks, got 0")
	}

	if checks[0].Status != StatusPass {
		t.Errorf("expected StatusPass for identical env keys, got %s: %s", checks[0].Status, checks[0].Message)
	}
}

func TestCheckEnvironmentVariables_MissingKeys(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, ".env.example"), []byte("PORT=8080\nDB_URL=postgres://...\nJWT_SECRET=abc"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("PORT=8080"), 0644)

	checks := checkEnvironmentVariables(tempDir)
	if len(checks) == 0 {
		t.Fatalf("expected checks, got 0")
	}

	if checks[0].Status != StatusFail {
		t.Errorf("expected StatusFail when variables are missing, got %s", checks[0].Status)
	}
	if len(checks[0].Details) != 2 {
		t.Errorf("expected 2 missing keys (DB_URL, JWT_SECRET), got %d: %v", len(checks[0].Details), checks[0].Details)
	}
}

func TestCheckGitAndSecrets_LeakProtection(t *testing.T) {
	tempDir := t.TempDir()

	// Mock git dir
	_ = os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)

	// .gitignore without .env
	_ = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("node_modules\ndist"), 0644)

	checks := checkGitAndSecrets(tempDir)
	foundRisk := false
	for _, c := range checks {
		if c.Category == "Security" && c.Status == StatusFail {
			foundRisk = true
			break
		}
	}
	if !foundRisk {
		t.Errorf("expected security failure when .env is omitted from .gitignore")
	}

	// Now fix .gitignore
	_ = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("node_modules\n.env\n.env.local"), 0644)
	checks2 := checkGitAndSecrets(tempDir)
	for _, c := range checks2 {
		if c.Category == "Security" && c.Status != StatusPass {
			t.Errorf("expected security pass when .env is ignored, got: %s", c.Status)
		}
	}
}

func TestRunAudit_FullProject(t *testing.T) {
	tempDir := t.TempDir()

	// Setup clean Go project
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "go.sum"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env.example"), []byte("PORT=9999"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("PORT=9999"), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(".env\nbin/\n"), 0644)

	report, err := RunAudit(AuditOptions{
		TargetDir:    tempDir,
		CheckNetwork: true,
	})
	if err != nil {
		t.Fatalf("RunAudit failed: %v", err)
	}

	if report.HealthScore < 90 {
		t.Errorf("expected high health score for clean project, got %d%%", report.HealthScore)
	}
	if report.Language != "go" {
		t.Errorf("expected language go, got %s", report.Language)
	}

	// Test ToJSON
	jsonBytes, err := report.ToJSON(true)
	if err != nil {
		t.Fatalf("failed serializing to JSON: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Errorf("expected non-empty JSON")
	}

	// Test RenderReportTo
	var buf bytes.Buffer
	RenderReportTo(&buf, report, false)
	if !strings.Contains(buf.String(), "UMARU PROJECT AUDIT") {
		t.Errorf("rendered output missing header: %s", buf.String())
	}
}

func TestRunAudit_WithLinter(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("PORT=8080"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env.example"), []byte("PORT=8080"), 0644)

	report, err := RunAudit(AuditOptions{
		TargetDir:   tempDir,
		CheckLinter: true,
	})
	if err != nil {
		t.Fatalf("RunAudit with linter failed: %v", err)
	}
	if report == nil {
		t.Fatalf("expected non-nil report")
	}
}

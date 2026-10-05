package audit

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"umaru/internal/generator"
)

// CheckStatus represents the outcome of an individual audit check
type CheckStatus string

const (
	StatusPass CheckStatus = "pass"
	StatusWarn CheckStatus = "warn"
	StatusFail CheckStatus = "fail"
)

// AuditCheck represents a single health or security check
type AuditCheck struct {
	Category string      `json:"category"` // "Environment", "Security", "Networking", "Dependencies", "Git"
	Title    string      `json:"title"`
	Status   CheckStatus `json:"status"`
	Message  string      `json:"message"`
	Details  []string    `json:"details,omitempty"`
}

// AuditReport aggregates all checks and computes a project health score
type AuditReport struct {
	ProjectName     string       `json:"project_name"`
	TargetDir       string       `json:"target_dir"`
	Language        string       `json:"language"`
	Framework       string       `json:"framework"`
	HealthScore     int          `json:"health_score"` // 0-100
	Timestamp       time.Time    `json:"timestamp"`
	Checks          []AuditCheck `json:"checks"`
	Recommendations []string     `json:"recommendations"`
}

// ToJSON serializes the audit report into JSON format
func (r *AuditReport) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

// AuditOptions configures the auditor
type AuditOptions struct {
	TargetDir    string
	CheckNetwork bool // Test port availability
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// RunAudit performs end-to-end health and security diagnostics on a project
func RunAudit(opts AuditOptions) (*AuditReport, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed resolving directory: %w", err)
	}

	proj, err := generator.DetectProject(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed detecting project: %w", err)
	}

	report := &AuditReport{
		ProjectName: proj.ProjectName,
		TargetDir:   absDir,
		Language:    string(proj.Type),
		Framework:   proj.Framework,
		Timestamp:   time.Now(),
		Checks:      make([]AuditCheck, 0),
	}

	baseDir := generator.GetAddonBaseDir(proj.ToProjectConfig(generator.AddonConfig{}))

	// 1. Environment Variables Audit (.env vs .env.example)
	report.Checks = append(report.Checks, checkEnvironmentVariables(baseDir)...)

	// 2. Git & Secret Leak Protection (.gitignore & .env leaks)
	report.Checks = append(report.Checks, checkGitAndSecrets(absDir)...)

	// 3. Dependency Installation Check
	report.Checks = append(report.Checks, checkDependencies(absDir, proj)...)

	// 4. Port Availability Check
	if opts.CheckNetwork {
		report.Checks = append(report.Checks, checkPortAvailability(baseDir)...)
	}

	// Calculate overall health score and extract recommendations
	calculateScoreAndRecommendations(report)

	return report, nil
}

// checkEnvironmentVariables inspects .env and .env.example consistency
func checkEnvironmentVariables(baseDir string) []AuditCheck {
	var checks []AuditCheck

	examplePath := filepath.Join(baseDir, ".env.example")
	if !fileExists(examplePath) {
		examplePath = filepath.Join(baseDir, ".env.sample")
	}
	if !fileExists(examplePath) {
		examplePath = filepath.Join(baseDir, ".env.template")
	}

	envPath := filepath.Join(baseDir, ".env")

	if !fileExists(examplePath) && !fileExists(envPath) {
		checks = append(checks, AuditCheck{
			Category: "Environment",
			Title:    "Environment Configuration",
			Status:   StatusWarn,
			Message:  "No .env or .env.example file found in project root",
			Details:  []string{"Consider creating a .env.example file to document required environment variables"},
		})
		return checks
	}

	if fileExists(examplePath) && !fileExists(envPath) {
		checks = append(checks, AuditCheck{
			Category: "Environment",
			Title:    "Local Environment File",
			Status:   StatusFail,
			Message:  "Missing local .env file (found .env.example)",
			Details:  []string{fmt.Sprintf("Run 'cp %s %s' and fill in your local secrets", filepath.Base(examplePath), filepath.Base(envPath))},
		})
		return checks
	}

	if !fileExists(examplePath) && fileExists(envPath) {
		checks = append(checks, AuditCheck{
			Category: "Environment",
			Title:    "Template Environment File",
			Status:   StatusWarn,
			Message:  "Found .env but missing .env.example template",
			Details:  []string{"Create a .env.example to help collaborators configure their environment"},
		})
		return checks
	}

	// Both files exist: compare variable keys
	exampleKeys := parseEnvKeys(examplePath)
	envKeys := parseEnvKeys(envPath)

	var missingKeys []string
	for k := range exampleKeys {
		if _, exists := envKeys[k]; !exists {
			missingKeys = append(missingKeys, k)
		}
	}

	if len(missingKeys) > 0 {
		checks = append(checks, AuditCheck{
			Category: "Environment",
			Title:    "Environment Variables Parity",
			Status:   StatusFail,
			Message:  fmt.Sprintf("%d variable(s) present in .env.example are missing from your local .env", len(missingKeys)),
			Details:  missingKeys,
		})
	} else {
		checks = append(checks, AuditCheck{
			Category: "Environment",
			Title:    "Environment Variables Parity",
			Status:   StatusPass,
			Message:  "All variables from .env.example are configured in your local .env",
		})
	}

	return checks
}

func parseEnvKeys(path string) map[string]string {
	keys := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return keys
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		var val string
		if len(parts) > 1 {
			val = strings.TrimSpace(parts[1])
		}
		if key != "" {
			keys[key] = val
		}
	}
	return keys
}

// checkGitAndSecrets ensures git is initialized and secrets are not leaked
func checkGitAndSecrets(dir string) []AuditCheck {
	var checks []AuditCheck

	gitDir := filepath.Join(dir, ".git")
	if !dirExists(gitDir) {
		checks = append(checks, AuditCheck{
			Category: "Git",
			Title:    "Git Repository",
			Status:   StatusWarn,
			Message:  "Project is not a Git repository",
			Details:  []string{"Run 'git init' to enable version control and track changes"},
		})
		return checks
	}

	checks = append(checks, AuditCheck{
		Category: "Git",
		Title:    "Git Repository",
		Status:   StatusPass,
		Message:  "Git repository is properly initialized",
	})

	// Check .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if !fileExists(gitignorePath) {
		checks = append(checks, AuditCheck{
			Category: "Security",
			Title:    "Gitignore Protection",
			Status:   StatusFail,
			Message:  "Missing .gitignore file (risk of leaking secrets and build artifacts)",
			Details:  []string{"Create a .gitignore to exclude .env and vendor directories"},
		})
		return checks
	}

	data, _ := os.ReadFile(gitignorePath)
	content := string(data)
	var securityRisks []string

	if !strings.Contains(content, ".env") {
		securityRisks = append(securityRisks, "'.env' is not listed in .gitignore (HIGH RISK of accidental secret commit)")
	}

	if len(securityRisks) > 0 {
		checks = append(checks, AuditCheck{
			Category: "Security",
			Title:    "Secret Leak Protection",
			Status:   StatusFail,
			Message:  "Sensitive files are not excluded by .gitignore",
			Details:  securityRisks,
		})
	} else {
		checks = append(checks, AuditCheck{
			Category: "Security",
			Title:    "Secret Leak Protection",
			Status:   StatusPass,
			Message:  ".env files are safely ignored in .gitignore",
		})
	}

	return checks
}

// checkDependencies verifies whether required runtime packages or lockfiles exist
func checkDependencies(dir string, proj *generator.DetectedProject) []AuditCheck {
	var checks []AuditCheck

	switch proj.Type {
	case generator.ProjectTypeNode:
		nodeModules := filepath.Join(dir, "node_modules")
		if !dirExists(nodeModules) {
			checks = append(checks, AuditCheck{
				Category: "Dependencies",
				Title:    "Node.js Dependencies",
				Status:   StatusFail,
				Message:  "node_modules folder is missing",
				Details:  []string{"Run 'npm install', 'pnpm install', or 'bun install' to install project dependencies"},
			})
		} else {
			checks = append(checks, AuditCheck{
				Category: "Dependencies",
				Title:    "Node.js Dependencies",
				Status:   StatusPass,
				Message:  "node_modules dependencies are installed",
			})
		}
	case generator.ProjectTypePython:
		venvDir := filepath.Join(dir, ".venv")
		if !dirExists(venvDir) && !dirExists(filepath.Join(dir, "venv")) {
			checks = append(checks, AuditCheck{
				Category: "Dependencies",
				Title:    "Python Virtual Environment",
				Status:   StatusWarn,
				Message:  "No local .venv directory detected",
				Details:  []string{"Run 'python -m venv .venv' or 'uv venv' to isolate dependencies"},
			})
		} else {
			checks = append(checks, AuditCheck{
				Category: "Dependencies",
				Title:    "Python Virtual Environment",
				Status:   StatusPass,
				Message:  "Local virtual environment (.venv) is present",
			})
		}
	case generator.ProjectTypeGo:
		goSumPath := filepath.Join(dir, "go.sum")
		if !fileExists(goSumPath) {
			checks = append(checks, AuditCheck{
				Category: "Dependencies",
				Title:    "Go Checksums",
				Status:   StatusWarn,
				Message:  "go.sum file not found",
				Details:  []string{"Run 'go mod tidy' to download dependencies and verify checksums"},
			})
		} else {
			checks = append(checks, AuditCheck{
				Category: "Dependencies",
				Title:    "Go Dependencies",
				Status:   StatusPass,
				Message:  "go.mod and go.sum are configured",
			})
		}
	}

	return checks
}

// checkPortAvailability checks if the application's configured port is currently free
func checkPortAvailability(baseDir string) []AuditCheck {
	var checks []AuditCheck

	port := "8080"
	envPath := filepath.Join(baseDir, ".env")
	if fileExists(envPath) {
		keys := parseEnvKeys(envPath)
		if p, ok := keys["PORT"]; ok && strings.TrimSpace(p) != "" {
			port = strings.TrimSpace(p)
		}
	}

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		checks = append(checks, AuditCheck{
			Category: "Networking",
			Title:    fmt.Sprintf("Port Availability (%s)", port),
			Status:   StatusWarn,
			Message:  fmt.Sprintf("Port %s is currently in use by another process", port),
			Details:  []string{fmt.Sprintf("Application may fail to start on %s or trigger an address collision", addr)},
		})
	} else {
		_ = l.Close()
		checks = append(checks, AuditCheck{
			Category: "Networking",
			Title:    fmt.Sprintf("Port Availability (%s)", port),
			Status:   StatusPass,
			Message:  fmt.Sprintf("Port %s is available for binding", port),
		})
	}

	return checks
}

func calculateScoreAndRecommendations(report *AuditReport) {
	if len(report.Checks) == 0 {
		report.HealthScore = 100
		return
	}

	passed := 0
	warnings := 0
	failures := 0

	for _, c := range report.Checks {
		switch c.Status {
		case StatusPass:
			passed++
		case StatusWarn:
			warnings++
			report.Recommendations = append(report.Recommendations, c.Details...)
		case StatusFail:
			failures++
			report.Recommendations = append(report.Recommendations, c.Details...)
		}
	}

	// Pass = 100%, Warn = 50%, Fail = 0%
	totalChecks := len(report.Checks)
	earned := (passed * 100) + (warnings * 50)
	report.HealthScore = earned / totalChecks
}

// RunLinterAudit runs available project linters if installed
func RunLinterAudit(ctx context.Context, dir string, projType generator.ProjectType) *AuditCheck {
	switch projType {
	case generator.ProjectTypeGo:
		if _, err := exec.LookPath("golangci-lint"); err == nil {
			cmd := exec.CommandContext(ctx, "golangci-lint", "run", "--timeout=1m")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err != nil {
				return &AuditCheck{
					Category: "Quality",
					Title:    "Golangci-lint Code Quality",
					Status:   StatusWarn,
					Message:  "Linter detected issues",
					Details:  strings.Split(strings.TrimSpace(string(out)), "\n"),
				}
			}
			return &AuditCheck{
				Category: "Quality",
				Title:    "Golangci-lint Code Quality",
				Status:   StatusPass,
				Message:  "All linter checks passed cleanly",
			}
		}
	}
	return nil
}

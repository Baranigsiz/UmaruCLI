package doctor

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"umaru/internal/templates"
)

// CheckStatus represents the status of a tool or check
type CheckStatus int

const (
	StatusOk CheckStatus = iota
	StatusWarning
	StatusMissing
)

// ToolCheck defines an individual tool or service to be inspected
type ToolCheck struct {
	Name        string
	Category    string
	Commands    []string // e.g. ["python", "python3"] in order of preference
	VersionArgs []string
	Required    bool
	Status      CheckStatus
	Version     string
	Path        string
	Description string
	Notes       string
	InstallTip  string
}

// SystemInfo holds host environment details
type SystemInfo struct {
	OS          string
	Arch        string
	NumCPU      int
	UmaruVer    string
}

// TemplateReadiness summarizes how many templates are ready to run
type TemplateReadiness struct {
	Category    string
	Total       int
	Ready       int
	Missing     []string
	IsReady     bool
}

// DoctorReport contains all collected diagnostic data
type DoctorReport struct {
	System      SystemInfo
	Tools       []ToolCheck
	Templates   []TemplateReadiness
	TotalScore  int // percentage readiness 0-100
}

// CleanVersion extracts a clean version string from raw CLI output
func CleanVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "unknown"
	}

	// Remove common prefixes
	lines := strings.Split(raw, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Regex to match semver or version numbers like 1.24.0 or v20.10.0 or 3.11.5
	re := regexp.MustCompile(`v?(\d+\.\d+(\.\d+)?(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?|\d+\.\d+)`)
	match := re.FindString(firstLine)
	if match != "" {
		return strings.TrimPrefix(match, "v")
	}

	// Fallback to first line if no semver pattern found
	if len(firstLine) > 30 {
		return firstLine[:30] + "..."
	}
	return firstLine
}

// RunDiagnostics collects all diagnostic data from the environment
func RunDiagnostics(umaruVersion string) DoctorReport {
	sys := SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		NumCPU:   runtime.NumCPU(),
		UmaruVer: umaruVersion,
	}

	toolDefs := []ToolCheck{
		// Version Control
		{
			Name:        "Git",
			Category:    "Version Control",
			Commands:    []string{"git"},
			VersionArgs: []string{"--version"},
			Required:    true,
			Description: "Required for repository initialization",
			InstallTip:  "https://git-scm.com/downloads",
		},
		// Core Programming Runtimes
		{
			Name:        "Go",
			Category:    "Runtimes",
			Commands:    []string{"go"},
			VersionArgs: []string{"version"},
			Required:    false,
			Description: "Required for Go Fiber, Gin, Echo & CLI templates",
			InstallTip:  "https://go.dev/dl/",
		},
		{
			Name:        "Node.js",
			Category:    "Runtimes",
			Commands:    []string{"node"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Required for Express, Fastify, Hono, NestJS & React",
			InstallTip:  "https://nodejs.org/",
		},
		{
			Name:        "Python",
			Category:    "Runtimes",
			Commands:    []string{"python", "python3"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Required for FastAPI starter templates",
			InstallTip:  "https://www.python.org/downloads/",
		},
		{
			Name:        "Cargo (Rust)",
			Category:    "Runtimes",
			Commands:    []string{"cargo"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Required for Actix, Axum & Rust CLI templates",
			InstallTip:  "https://rustup.rs/",
		},
		// Package Managers
		{
			Name:        "npm",
			Category:    "Package Managers",
			Commands:    []string{"npm"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Default Node.js package manager",
			InstallTip:  "Bundled with Node.js",
		},
		{
			Name:        "pnpm",
			Category:    "Package Managers",
			Commands:    []string{"pnpm"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Fast, disk space efficient package manager (recommended for monorepos)",
			InstallTip:  "npm install -g pnpm",
		},
		{
			Name:        "yarn",
			Category:    "Package Managers",
			Commands:    []string{"yarn"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Alternative package manager for JavaScript",
			InstallTip:  "npm install -g yarn",
		},
		{
			Name:        "bun",
			Category:    "Package Managers",
			Commands:    []string{"bun"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "All-in-one JavaScript runtime & toolkit",
			InstallTip:  "https://bun.sh/",
		},
		{
			Name:        "pip",
			Category:    "Package Managers",
			Commands:    []string{"pip", "pip3"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Python package installer",
			InstallTip:  "Bundled with Python",
		},
		// Containers & Orchestration
		{
			Name:        "Docker CLI",
			Category:    "Containers",
			Commands:    []string{"docker"},
			VersionArgs: []string{"--version"},
			Required:    false,
			Description: "Container CLI for local databases and deployments",
			InstallTip:  "https://www.docker.com/products/docker-desktop/",
		},
		{
			Name:        "Docker Compose",
			Category:    "Containers",
			Commands:    []string{"docker"},
			VersionArgs: []string{"compose", "version"},
			Required:    false,
			Description: "Multi-container orchestration for PostgreSQL/Redis addons",
			InstallTip:  "Included with Docker Desktop",
		},
	}

	results := make([]ToolCheck, len(toolDefs))
	var daemonStatus ToolCheck
	var wg sync.WaitGroup
	wg.Add(len(toolDefs) + 1)

	for i, tc := range toolDefs {
		go func(idx int, check ToolCheck) {
			defer wg.Done()
			results[idx] = inspectTool(check)
		}(i, tc)
	}

	go func() {
		defer wg.Done()
		daemonStatus = inspectDockerDaemon()
	}()

	wg.Wait()

	toolMap := make(map[string]ToolCheck, len(results)+1)
	for _, checked := range results {
		toolMap[strings.ToLower(checked.Name)] = checked
	}

	results = append(results, daemonStatus)
	toolMap["docker daemon"] = daemonStatus

	// Calculate Template Readiness Matrix
	readiness := calculateReadiness(toolMap)

	// Calculate readiness score
	readyCount := 0
	totalCount := 0
	for _, r := range readiness {
		readyCount += r.Ready
		totalCount += r.Total
	}

	score := 100
	if totalCount > 0 {
		score = (readyCount * 100) / totalCount
	}

	return DoctorReport{
		System:     sys,
		Tools:      results,
		Templates:  readiness,
		TotalScore: score,
	}
}

func inspectTool(tc ToolCheck) ToolCheck {
	var resolvedPath string
	var chosenCmd string

	for _, cmd := range tc.Commands {
		if path, err := exec.LookPath(cmd); err == nil {
			resolvedPath = path
			chosenCmd = cmd
			break
		}
	}

	if resolvedPath == "" {
		tc.Status = StatusMissing
		tc.Notes = "Not found in PATH"
		return tc
	}

	tc.Path = resolvedPath

	// Run version command with a 2-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, chosenCmd, tc.VersionArgs...)
	out, err := cmd.Output()
	if err != nil {
		tc.Status = StatusWarning
		tc.Notes = "Installed, but version check failed"
		return tc
	}

	tc.Status = StatusOk
	tc.Version = CleanVersion(string(out))
	return tc
}

func inspectDockerDaemon() ToolCheck {
	check := ToolCheck{
		Name:        "Docker Daemon",
		Category:    "Containers",
		Required:    false,
		Description: "Docker daemon engine state",
		InstallTip:  "Start Docker Desktop or run 'sudo systemctl start docker'",
	}

	if _, err := exec.LookPath("docker"); err != nil {
		check.Status = StatusMissing
		check.Notes = "Docker CLI not found"
		check.InstallTip = "Install Docker Desktop"
		return check
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ServerVersion}}")
	out, err := cmd.Output()
	if err != nil {
		check.Status = StatusWarning
		check.Notes = "Daemon is stopped or unresponsive"
		return check
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		version = "Active"
	}

	check.Status = StatusOk
	check.Version = CleanVersion(version)
	check.Notes = "Daemon is running"
	return check
}

func calculateReadiness(tools map[string]ToolCheck) []TemplateReadiness {
	allTemplates, err := templates.GetAvailableTemplates()
	if err != nil {
		return nil
	}

	hasGo := isOk(tools, "go")
	hasNode := isOk(tools, "node.js")
	hasNpm := isOk(tools, "npm") || isOk(tools, "pnpm") || isOk(tools, "yarn") || isOk(tools, "bun")
	hasPython := isOk(tools, "python")
	hasPip := isOk(tools, "pip")
	hasRust := isOk(tools, "cargo (rust)")

	groups := map[string]*TemplateReadiness{
		"Go": {
			Category: "Go (Fiber, Gin, Echo, Cobra CLI)",
		},
		"Node/TypeScript": {
			Category: "Node/TypeScript (Express, Fastify, Hono, NestJS, React, Vue, Svelte, Next, Astro)",
		},
		"Python": {
			Category: "Python (FastAPI)",
		},
		"Rust": {
			Category: "Rust (Actix, Axum)",
		},
		"Fullstack": {
			Category: "Fullstack Monorepos (Go+React, Hono+React)",
		},
	}

	for _, tmpl := range allTemplates {
		id := tmpl.ID
		switch {
		case strings.HasPrefix(id, "go-"):
			groups["Go"].Total++
			if hasGo {
				groups["Go"].Ready++
			} else {
				groups["Go"].Missing = appendUnique(groups["Go"].Missing, "go")
			}
		case strings.HasPrefix(id, "rust-"):
			groups["Rust"].Total++
			if hasRust {
				groups["Rust"].Ready++
			} else {
				groups["Rust"].Missing = appendUnique(groups["Rust"].Missing, "cargo")
			}
		case strings.HasPrefix(id, "python-") || id == "fastapi" || strings.HasPrefix(id, "ai-"):
			groups["Python"].Total++
			if hasPython && hasPip {
				groups["Python"].Ready++
			} else {
				if !hasPython {
					groups["Python"].Missing = appendUnique(groups["Python"].Missing, "python")
				}
				if !hasPip {
					groups["Python"].Missing = appendUnique(groups["Python"].Missing, "pip")
				}
			}
		case id == "fullstack-go-react":
			groups["Fullstack"].Total++
			if hasGo && hasNode && hasNpm {
				groups["Fullstack"].Ready++
			} else {
				if !hasGo {
					groups["Fullstack"].Missing = appendUnique(groups["Fullstack"].Missing, "go")
				}
				if !hasNode {
					groups["Fullstack"].Missing = appendUnique(groups["Fullstack"].Missing, "node")
				}
			}
		case id == "fullstack-ts-monorepo":
			groups["Fullstack"].Total++
			if hasNode && hasNpm {
				groups["Fullstack"].Ready++
			} else {
				if !hasNode {
					groups["Fullstack"].Missing = appendUnique(groups["Fullstack"].Missing, "node")
				}
			}
		case id == "bun-elysia":
			groups["Node/TypeScript"].Total++
			if isOk(tools, "bun") || (hasNode && hasNpm) {
				groups["Node/TypeScript"].Ready++
			} else {
				groups["Node/TypeScript"].Missing = appendUnique(groups["Node/TypeScript"].Missing, "bun")
			}
		default: // All frontend & node backend
			groups["Node/TypeScript"].Total++
			if hasNode && hasNpm {
				groups["Node/TypeScript"].Ready++
			} else {
				if !hasNode {
					groups["Node/TypeScript"].Missing = appendUnique(groups["Node/TypeScript"].Missing, "node")
				}
				if !hasNpm {
					groups["Node/TypeScript"].Missing = appendUnique(groups["Node/TypeScript"].Missing, "npm/pnpm")
				}
			}
		}
	}

	// Order results nicely
	order := []string{"Go", "Node/TypeScript", "Python", "Rust", "Fullstack"}
	var list []TemplateReadiness
	for _, key := range order {
		if g, exists := groups[key]; exists && g.Total > 0 {
			g.IsReady = (g.Ready == g.Total)
			list = append(list, *g)
		}
	}

	return list
}

func isOk(tools map[string]ToolCheck, key string) bool {
	t, ok := tools[key]
	return ok && t.Status == StatusOk
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

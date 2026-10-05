package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"umaru/internal/generator"
)

// TargetPlatform represents supported cloud deployment providers
type TargetPlatform string

const (
	PlatformFly     TargetPlatform = "fly"
	PlatformRailway TargetPlatform = "railway"
	PlatformRender  TargetPlatform = "render"
	PlatformDocker  TargetPlatform = "docker"
)

// SupportedPlatforms returns the list of available deployment targets
func SupportedPlatforms() []TargetPlatform {
	return []TargetPlatform{
		PlatformFly,
		PlatformRailway,
		PlatformRender,
		PlatformDocker,
	}
}

// DeployOptions configures the deployment generator
type DeployOptions struct {
	TargetDir string         `json:"target_dir"`
	Platform  TargetPlatform `json:"platform"`
	AppName   string         `json:"app_name"`
	Port      int            `json:"port"`
	Force     bool           `json:"force"`
	DryRun    bool           `json:"dry_run"`
}

// GeneratedFile represents a deployment configuration file
type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Action  string `json:"action"` // "created", "overwritten", "skipped"
}

// DeployResult captures all generated files and deployment instructions
type DeployResult struct {
	Platform     TargetPlatform  `json:"platform"`
	AppName      string          `json:"app_name"`
	Port         int             `json:"port"`
	TargetDir    string          `json:"target_dir"`
	Files        []GeneratedFile `json:"files"`
	Instructions []string        `json:"instructions"`
}

// ToJSON serializes the deployment result to formatted JSON
func (r *DeployResult) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// GenerateDeployment inspects the project and emits platform-specific deployment manifests
func GenerateDeployment(opts DeployOptions) (*DeployResult, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed resolving target directory: %w", err)
	}

	proj, err := generator.DetectProject(absDir)
	if err != nil {
		return nil, fmt.Errorf("project detection failed: %w", err)
	}

	appName := opts.AppName
	if appName == "" {
		appName = generator.Slugify(proj.ProjectName)
		if appName == "" {
			appName = "my-app"
		}
	}

	port := opts.Port
	if port <= 0 {
		port = defaultPortForProject(proj)
	}

	opts.AppName = appName
	opts.Port = port
	opts.TargetDir = absDir

	var files []GeneratedFile
	var instructions []string

	switch strings.ToLower(string(opts.Platform)) {
	case string(PlatformFly), "fly.io":
		files, instructions = generateFly(proj, opts)
	case string(PlatformRailway):
		files, instructions = generateRailway(proj, opts)
	case string(PlatformRender):
		files, instructions = generateRender(proj, opts)
	case string(PlatformDocker), "dockerfile":
		files, instructions = generateDocker(proj, opts)
	default:
		return nil, fmt.Errorf("unsupported deployment platform '%s'. Available targets: fly, railway, render, docker", opts.Platform)
	}

	// Write files if not dry-run
	for i := range files {
		f := &files[i]
		fullPath := filepath.Join(absDir, f.Path)
		alreadyExists := fileExists(fullPath)

		if alreadyExists && !opts.Force {
			f.Action = "skipped (exists)"
			continue
		}

		if opts.DryRun {
			if alreadyExists {
				f.Action = "overwrite (dry-run)"
			} else {
				f.Action = "create (dry-run)"
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return nil, fmt.Errorf("failed creating directory for %s: %w", f.Path, err)
		}

		if err := os.WriteFile(fullPath, []byte(f.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed writing %s: %w", f.Path, err)
		}

		if alreadyExists {
			f.Action = "overwritten"
		} else {
			f.Action = "created"
		}
	}

	return &DeployResult{
		Platform:     opts.Platform,
		AppName:      appName,
		Port:         port,
		TargetDir:    absDir,
		Files:        files,
		Instructions: instructions,
	}, nil
}

func defaultPortForProject(proj *generator.DetectedProject) int {
	switch proj.Type {
	case generator.ProjectTypeGo:
		return 8080
	case generator.ProjectTypeNode:
		return 3000
	case generator.ProjectTypePython:
		return 8000
	case generator.ProjectTypeRust:
		return 8080
	default:
		return 8080
	}
}

func generateFly(_ *generator.DetectedProject, opts DeployOptions) ([]GeneratedFile, []string) {
	content := fmt.Sprintf(`# fly.toml app configuration file generated for %s by Umaru CLI
# See https://fly.io/docs/reference/configuration/ for reference

app = "%s"
primary_region = "fra"

[build]

[http_service]
  internal_port = %d
  force_https = true
  auto_stop_machines = "stop"
  auto_start_machines = true
  min_machines_running = 0
  processes = ["app"]

[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"
`, opts.AppName, opts.AppName, opts.Port)

	files := []GeneratedFile{
		{
			Path:    "fly.toml",
			Content: content,
		},
	}

	instructions := []string{
		"Install the Fly.io CLI: https://fly.io/docs/hands-on/install-flyctl/",
		"Authenticate your account: fly auth login",
		fmt.Sprintf("Deploy your app to Fly.io: fly deploy (uses internal port %d)", opts.Port),
		"View live logs: fly logs",
	}

	return files, instructions
}

func generateRailway(proj *generator.DetectedProject, opts DeployOptions) ([]GeneratedFile, []string) {
	builder := "NIXPACKS"
	startCmd := ""

	switch proj.Type {
	case generator.ProjectTypeGo:
		startCmd = "go run main.go"
	case generator.ProjectTypeNode:
		startCmd = "npm run start"
	case generator.ProjectTypePython:
		startCmd = fmt.Sprintf("uvicorn app.main:app --host 0.0.0.0 --port %d", opts.Port)
	case generator.ProjectTypeRust:
		startCmd = "cargo run --release"
	}

	configObj := map[string]interface{}{
		"$schema": "https://railway.com/railway.schema.json",
		"build": map[string]string{
			"builder": builder,
		},
		"deploy": map[string]interface{}{
			"startCommand":            startCmd,
			"healthcheckPath":         "/",
			"healthcheckTimeout":      100,
			"restartPolicyType":       "ON_FAILURE",
			"restartPolicyMaxRetries": 10,
		},
	}

	data, _ := json.MarshalIndent(configObj, "", "  ")

	files := []GeneratedFile{
		{
			Path:    "railway.json",
			Content: string(data) + "\n",
		},
	}

	instructions := []string{
		"Install Railway CLI: npm install -g @railway/cli (or brew install railway)",
		"Link to your Railway project: railway link",
		"Deploy your project to Railway: railway up",
		"Open your dashboard: railway open",
	}

	return files, instructions
}

func generateRender(proj *generator.DetectedProject, opts DeployOptions) ([]GeneratedFile, []string) {
	runtime := "docker"
	buildCmd := ""
	startCmd := ""

	switch proj.Type {
	case generator.ProjectTypeGo:
		runtime = "go"
		buildCmd = "go build -o server main.go"
		startCmd = "./server"
	case generator.ProjectTypeNode:
		runtime = "node"
		buildCmd = "npm install && npm run build --if-present"
		startCmd = "npm run start"
	case generator.ProjectTypePython:
		runtime = "python"
		buildCmd = "pip install -r requirements.txt"
		startCmd = fmt.Sprintf("uvicorn app.main:app --host 0.0.0.0 --port %d", opts.Port)
	case generator.ProjectTypeRust:
		runtime = "docker"
	}

	content := fmt.Sprintf(`services:
  - type: web
    name: %s
    runtime: %s
    plan: free
    region: frankfurt
`, opts.AppName, runtime)

	if buildCmd != "" {
		content += fmt.Sprintf("    buildCommand: \"%s\"\n", buildCmd)
	}
	if startCmd != "" {
		content += fmt.Sprintf("    startCommand: \"%s\"\n", startCmd)
	}

	content += fmt.Sprintf(`    envVars:
      - key: PORT
        value: %d
`, opts.Port)

	files := []GeneratedFile{
		{
			Path:    "render.yaml",
			Content: content,
		},
	}

	instructions := []string{
		"Push this repository to GitHub or GitLab",
		"Sign in to your Render dashboard: https://dashboard.render.com",
		"Click 'New' -> 'Blueprint' and select this repository",
		"Render will automatically provision your web service from render.yaml",
	}

	return files, instructions
}

func generateDocker(proj *generator.DetectedProject, opts DeployOptions) ([]GeneratedFile, []string) {
	var content string

	switch proj.Type {
	case generator.ProjectTypeGo:
		content = fmt.Sprintf(`# Production Multi-Stage Dockerfile generated for %s by Umaru CLI
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server .

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE %d
USER nonroot:nonroot
CMD ["/app/server"]
`, opts.AppName, opts.Port)

	case generator.ProjectTypeNode:
		content = fmt.Sprintf(`# Production Multi-Stage Dockerfile generated for %s by Umaru CLI
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json pnpm-lock.yaml* yarn.lock* bun.lock* ./
RUN npm ci || npm install
COPY . .
RUN npm run build --if-present

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY package*.json ./
RUN npm ci --only=production || npm install --production
COPY --from=builder /app/dist ./dist
EXPOSE %d
USER node
CMD ["node", "dist/index.js"]
`, opts.AppName, opts.Port)

	case generator.ProjectTypePython:
		content = fmt.Sprintf(`# Production Dockerfile generated for %s by Umaru CLI
FROM python:3.12-slim AS runner
WORKDIR /app
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE %d
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "%d"]
`, opts.AppName, opts.Port, opts.Port)

	default:
		content = fmt.Sprintf(`# Production Dockerfile generated for %s by Umaru CLI
FROM alpine:latest
WORKDIR /app
COPY . .
EXPOSE %d
CMD ["./server"]
`, opts.AppName, opts.Port)
	}

	files := []GeneratedFile{
		{
			Path:    "Dockerfile.prod",
			Content: content,
		},
	}

	instructions := []string{
		fmt.Sprintf("Build your production Docker image: docker build -f Dockerfile.prod -t %s:latest .", opts.AppName),
		fmt.Sprintf("Run the container locally: docker run -p %d:%d %s:latest", opts.Port, opts.Port, opts.AppName),
		"Deploy image to Docker Hub, AWS ECR, or Google Artifact Registry",
	}

	return files, instructions
}

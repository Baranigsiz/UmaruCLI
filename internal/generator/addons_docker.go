package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getDockerFiles(baseDir string) []string {
	return []string{
		filepath.Join(baseDir, "Dockerfile"),
		filepath.Join(baseDir, "docker-compose.yml"),
		filepath.Join(baseDir, ".dockerignore"),
	}
}

func buildDockerCompose(config ProjectConfig, appService, appPort string, defaultEnv []string) string {
	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\nservices:\n")
	sb.WriteString(fmt.Sprintf("  %s:\n", appService))
	sb.WriteString("    build:\n      context: .\n      dockerfile: Dockerfile\n")
	sb.WriteString(fmt.Sprintf("    ports:\n      - \"%s:%s\"\n", appPort, appPort))

	hasPG := config.Addons.Database == "postgres"
	hasRedis := config.Addons.Redis

	var envList []string
	envList = append(envList, defaultEnv...)
	if hasPG {
		envList = append(envList,
			"DB_HOST=postgres",
			"DB_PORT=5432",
			"DB_USER=postgres",
			"DB_PASSWORD=postgres",
			fmt.Sprintf("DB_NAME=%s", config.SafeName),
		)
	}
	if hasRedis {
		envList = append(envList,
			"REDIS_ADDR=redis:6379",
			"REDIS_HOST=redis",
			"REDIS_PORT=6379",
		)
	}

	if len(envList) > 0 {
		sb.WriteString("    environment:\n")
		for _, env := range envList {
			sb.WriteString(fmt.Sprintf("      - %s\n", env))
		}
	}

	var dependsOn []string
	if hasPG {
		dependsOn = append(dependsOn, "postgres")
	}
	if hasRedis {
		dependsOn = append(dependsOn, "redis")
	}
	if len(dependsOn) > 0 {
		sb.WriteString("    depends_on:\n")
		for _, dep := range dependsOn {
			sb.WriteString(fmt.Sprintf("      - %s\n", dep))
		}
	}

	sb.WriteString("    restart: unless-stopped\n")

	var volumes []string

	if hasPG {
		sb.WriteString("\n  postgres:\n")
		sb.WriteString("    image: postgres:16-alpine\n")
		sb.WriteString(fmt.Sprintf("    container_name: %s-postgres\n", config.SafeName))
		sb.WriteString("    environment:\n")
		sb.WriteString("      POSTGRES_USER: postgres\n")
		sb.WriteString("      POSTGRES_PASSWORD: postgres\n")
		sb.WriteString(fmt.Sprintf("      POSTGRES_DB: %s\n", config.SafeName))
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"5432:5432\"\n")
		sb.WriteString(fmt.Sprintf("    volumes:\n      - %s_pgdata:/var/lib/postgresql/data\n", config.SafeName))
		sb.WriteString("    restart: unless-stopped\n")
		volumes = append(volumes, fmt.Sprintf("%s_pgdata", config.SafeName))
	}

	if hasRedis {
		sb.WriteString("\n  redis:\n")
		sb.WriteString("    image: redis:7-alpine\n")
		sb.WriteString(fmt.Sprintf("    container_name: %s-redis\n", config.SafeName))
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"6379:6379\"\n")
		sb.WriteString(fmt.Sprintf("    volumes:\n      - %s_redisdata:/data\n", config.SafeName))
		sb.WriteString("    restart: unless-stopped\n")
		volumes = append(volumes, fmt.Sprintf("%s_redisdata", config.SafeName))
	}

	if len(volumes) > 0 {
		sb.WriteString("\nvolumes:\n")
		for _, v := range volumes {
			sb.WriteString(fmt.Sprintf("  %s:\n", v))
		}
	}

	return sb.String()
}

var (
	topLevelSectionRegex = regexp.MustCompile(`(?m)^([a-zA-Z0-9_-]+):\s*$`)
	topLevelVolumesRegex = regexp.MustCompile(`(?m)^volumes:\s*$`)
)

func appendDockerComposeServices(baseDir string, config ProjectConfig) error {
	composePath := filepath.Join(baseDir, "docker-compose.yml")
	if !fileExists(composePath) && fileExists(filepath.Join(config.TargetDir, "docker-compose.yml")) {
		composePath = filepath.Join(config.TargetDir, "docker-compose.yml")
	}

	data, err := os.ReadFile(composePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	var toAppend strings.Builder
	var volumesToAppend []string

	hasPG := config.Addons.Database == "postgres"
	hasRedis := config.Addons.Redis

	if hasPG && !strings.Contains(content, "postgres:") {
		toAppend.WriteString(fmt.Sprintf(`  postgres:
    image: postgres:16-alpine
    container_name: %s-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: %s
    ports:
      - "5432:5432"
    volumes:
      - %s_pgdata:/var/lib/postgresql/data
    restart: unless-stopped
`, config.SafeName, config.SafeName, config.SafeName))
		volumesToAppend = append(volumesToAppend, fmt.Sprintf("%s_pgdata", config.SafeName))
	}

	if hasRedis && !strings.Contains(content, "redis:") {
		toAppend.WriteString(fmt.Sprintf(`  redis:
    image: redis:7-alpine
    container_name: %s-redis
    ports:
      - "6379:6379"
    volumes:
      - %s_redisdata:/data
    restart: unless-stopped
`, config.SafeName, config.SafeName))
		volumesToAppend = append(volumesToAppend, fmt.Sprintf("%s_redisdata", config.SafeName))
	}

	var dependsOn []string
	if hasPG {
		dependsOn = append(dependsOn, "postgres")
	}
	if hasRedis {
		dependsOn = append(dependsOn, "redis")
	}

	if len(dependsOn) > 0 {
		if !strings.Contains(content, "depends_on:") {
			var depBlock strings.Builder
			depBlock.WriteString("    depends_on:\n")
			for _, dep := range dependsOn {
				depBlock.WriteString(fmt.Sprintf("      - %s\n", dep))
			}
			if strings.Contains(content, "restart: unless-stopped") {
				content = strings.Replace(content, "restart: unless-stopped", depBlock.String()+"    restart: unless-stopped", 1)
			} else if strings.Contains(content, "ports:\n") {
				content = strings.Replace(content, "ports:\n", depBlock.String()+"    ports:\n", 1)
			} else if strings.Contains(content, "build:\n") {
				content = strings.Replace(content, "build:\n", depBlock.String()+"    build:\n", 1)
			}
		} else {
			for _, dep := range dependsOn {
				depEntry := "- " + dep
				if !strings.Contains(content, depEntry) {
					content = strings.Replace(content, "depends_on:\n", fmt.Sprintf("depends_on:\n      - %s\n", dep), 1)
				}
			}
		}
	}

	if toAppend.Len() == 0 && len(volumesToAppend) == 0 {
		return nil
	}

	// Insert new services before any top-level key following services: (volumes:, networks:, etc.)
	if toAppend.Len() > 0 {
		matches := topLevelSectionRegex.FindAllStringSubmatchIndex(content, -1)
		insertIdx := -1
		for _, m := range matches {
			secName := content[m[2]:m[3]]
			if secName != "version" && secName != "services" {
				insertIdx = m[0]
				break
			}
		}

		if insertIdx != -1 {
			prefix := strings.TrimRight(content[:insertIdx], "\n")
			suffix := strings.TrimLeft(content[insertIdx:], "\n")
			content = prefix + "\n\n" + toAppend.String() + "\n" + suffix
		} else {
			content = strings.TrimRight(content, "\n") + "\n\n" + toAppend.String()
		}
	}

	// Insert missing volumes under top-level volumes:
	if len(volumesToAppend) > 0 {
		vLoc := topLevelVolumesRegex.FindStringIndex(content)
		if vLoc != nil {
			var newVols strings.Builder
			for _, v := range volumesToAppend {
				if !strings.Contains(content, v+":") {
					newVols.WriteString(fmt.Sprintf("  %s:\n", v))
				}
			}
			if newVols.Len() > 0 {
				vEnd := vLoc[1]
				if vEnd < len(content) && content[vEnd] == '\n' {
					vEnd++
				}
				content = content[:vEnd] + newVols.String() + content[vEnd:]
			}
		} else {
			var newVols strings.Builder
			newVols.WriteString("\nvolumes:\n")
			for _, v := range volumesToAppend {
				newVols.WriteString(fmt.Sprintf("  %s:\n", v))
			}
			content = strings.TrimRight(content, "\n") + "\n" + newVols.String()
		}
	}

	return os.WriteFile(composePath, []byte(content), 0644)
}

func generateDockerAddon(config ProjectConfig, baseDir string) error {
	var dockerfileContent string
	var composeContent string
	dockerignoreContent := `.git
.gitignore
node_modules
dist
bin
tmp
target
__pycache__
*.pyc
.env
.env.*
.DS_Store
`

	switch {
	case isGoTemplate(config.Template):
		entrypoint := "."
		if fileExists(filepath.Join(baseDir, "cmd", "api", "main.go")) ||
			fileExists(filepath.Join(config.TargetDir, "cmd", "api", "main.go")) ||
			config.Template == "go-fiber" ||
			config.Template == "go-gin" ||
			config.Template == "go-echo" ||
			config.Template == "fullstack-go-react" {
			entrypoint = "cmd/api/main.go"
		} else if fileExists(filepath.Join(baseDir, "cmd", "web", "main.go")) ||
			fileExists(filepath.Join(config.TargetDir, "cmd", "web", "main.go")) ||
			config.Template == "go-htmx" {
			entrypoint = "cmd/web/main.go"
		}

		port := "8080"
		if config.Template == "go-fiber" || config.Template == "go-htmx" {
			port = "3000"
		}

		dockerfileContent = fmt.Sprintf(`# Multi-Stage Dockerfile for Go
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server %s

FROM alpine:3.20 AS runner
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server .
EXPOSE %s
ENTRYPOINT ["./server"]
`, entrypoint, port)
		composeContent = buildDockerCompose(config, "app", port, nil)

	case strings.HasPrefix(config.Template, "bun-"):
		dockerfileContent = `# Multi-Stage Dockerfile for Bun
FROM oven/bun:1 AS base
WORKDIR /app

FROM base AS install
COPY package.json bun.lock* ./
RUN bun install

FROM base AS release
COPY --from=install /app/node_modules ./node_modules
COPY . .

EXPOSE 3000
USER bun
ENTRYPOINT ["bun", "run", "src/index.ts"]
`
		composeContent = buildDockerCompose(config, "api", "3000", []string{"NODE_ENV=production", "PORT=3000"})

	case isPythonTemplate(config.Template):
		dockerfileContent = `# Dockerfile for Python
FROM python:3.11-slim AS base
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY . .
EXPOSE 8000
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
`
		composeContent = buildDockerCompose(config, "api", "8000", []string{"PORT=8000"})

	case isRustTemplate(config.Template):
		dockerfileContent = fmt.Sprintf(`# Multi-Stage Dockerfile for Rust
FROM rust:1.80-slim-bullseye AS builder
WORKDIR /app

COPY Cargo.toml ./
# Create dummy main to cache dependencies
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src

COPY src ./src
# Touch main.rs to invalidate the build timestamp
RUN touch src/main.rs && cargo build --release

# Production Stage
FROM debian:bullseye-slim AS runner
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/%s /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
`, config.SafeName)
		composeContent = buildDockerCompose(config, "app", "8080", nil)

	default: // Node.js / TypeScript
		dockerfileContent = `# Multi-Stage Dockerfile for Node.js
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* pnpm-lock.yaml* yarn.lock* ./
RUN npm install

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build --if-present

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app ./
EXPOSE 3000
CMD ["npm", "start"]
`
		composeContent = buildDockerCompose(config, "app", "3000", []string{"NODE_ENV=production", "PORT=3000"})
	}

	dockerfilePath := filepath.Join(baseDir, "Dockerfile")
	if !fileExists(dockerfilePath) {
		if err := writeAddonFile(baseDir, "Dockerfile", dockerfileContent); err != nil {
			return err
		}
	}

	composePath := filepath.Join(baseDir, "docker-compose.yml")
	if fileExists(composePath) {
		if err := appendDockerComposeServices(baseDir, config); err != nil {
			return err
		}
	} else {
		if err := writeAddonFile(baseDir, "docker-compose.yml", composeContent); err != nil {
			return err
		}
	}

	if err := writeAddonFile(baseDir, ".dockerignore", dockerignoreContent); err != nil {
		return err
	}

	return nil
}

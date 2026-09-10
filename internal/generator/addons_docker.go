package generator

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getDockerFiles(baseDir string) []string {
	return []string{
		filepath.Join(baseDir, "Dockerfile"),
		filepath.Join(baseDir, "docker-compose.yml"),
		filepath.Join(baseDir, ".dockerignore"),
	}
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
		dockerfileContent = `# Multi-Stage Dockerfile for Go
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

FROM alpine:3.20 AS runner
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server .
EXPOSE 8080
ENTRYPOINT ["./server"]
`
		composeContent = `version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    restart: unless-stopped
`
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
		composeContent = `version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - PORT=3000
    restart: unless-stopped
`
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
		composeContent = `version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8000:8000"
    environment:
      - PORT=8000
    restart: unless-stopped
`
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

		composeContent = `version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    restart: unless-stopped
`
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
		composeContent = `version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - PORT=3000
    restart: unless-stopped
`
	}

	if err := writeAddonFile(baseDir, "Dockerfile", dockerfileContent); err != nil {
		return err
	}
	if err := writeAddonFile(baseDir, "docker-compose.yml", composeContent); err != nil {
		return err
	}
	if err := writeAddonFile(baseDir, ".dockerignore", dockerignoreContent); err != nil {
		return err
	}

	return nil
}

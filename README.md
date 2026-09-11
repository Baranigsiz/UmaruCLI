<div align="center">

# ⚡ Umaru CLI

**A lightning-fast, production-grade project scaffolding CLI for modern developers.**

Bootstraps clean architecture backends, modern frontend apps, and monorepos in milliseconds — complete with Docker, Graceful Shutdown, OpenAPI, and interactive terminal UI.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![CI Workflow](https://img.shields.io/github/actions/workflow/status/Baranigsiz/UmaruCLI/ci.yml?branch=main&style=for-the-badge&label=CI&logo=githubactions&logoColor=white)](https://github.com/Baranigsiz/UmaruCLI/actions)
[![CodeQL Security](https://img.shields.io/github/actions/workflow/status/Baranigsiz/UmaruCLI/codeql.yml?branch=main&style=for-the-badge&label=CodeQL&logo=github&logoColor=white)](https://github.com/Baranigsiz/UmaruCLI/actions/workflows/codeql.yml)
[![Lint & Quality](https://img.shields.io/github/actions/workflow/status/Baranigsiz/UmaruCLI/lint.yml?branch=main&style=for-the-badge&label=Lint&logo=go&logoColor=white)](https://github.com/Baranigsiz/UmaruCLI/actions/workflows/lint.yml)
[![Release](https://img.shields.io/github/v/release/Baranigsiz/UmaruCLI?style=for-the-badge&logo=semanticrelease&logoColor=white&color=7D56F4)](https://github.com/Baranigsiz/UmaruCLI/releases)
[![License](https://img.shields.io/badge/License-MIT-emerald?style=for-the-badge)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge)](CONTRIBUTING.md)

<br />

<pre align="center">
  _   _                               ____ _     ___ 
 | | | |_ __ ___   __ _ _ __ _   _   / ___| |   |_ _|
 | | | | '_ ` _ \ / _` | '__| | | | | |   | |    | | 
 | |_| | | | | | | (_| | |  | |_| | | |___| |___ | | 
  \___/|_| |_| |_|\__,_|_|   \__,_|  \____|_____|___|
                                                     
          Production Scaffolding in Milliseconds
</pre>

<br />

<img src=".github/assets/demo.gif" alt="Umaru CLI Interactive Demo" width="850" />

<br /><br />

<p align="center">
  <a href="#-features">✨ Features</a> •
  <a href="#-real-world-scaffolding-benchmarks">⚡ Benchmarks</a> •
  <a href="#-supported-starters">📦 Starters (23)</a> •
  <a href="#-interactive-addon-wizard">🧩 Addon Wizard</a> •
  <a href="#️-global-configuration">⚙️ Global Config</a> •
  <a href="#-installation">🚀 Installation</a> •
  <a href="#-usage">💻 Usage</a> •
  <a href="#-system-diagnostics">🩺 Doctor</a> •
  <a href="#-template-deep-inspection-umaru-info">🔍 Inspect</a> •
  <a href="#-shell-autocompletion">🐚 Autocompletion</a> •
  <a href="#-self-upgrade">🔄 Self-Upgrade</a> •
  <a href="#️-extensibility--custom-templates">🛠️ Extensibility</a>
</p>

</div>

---

## 💡 Why Umaru?

Most scaffolding tools generate bare-bones, single-file "Hello World" scripts. When starting a real-world application, developers often spend hours configuring:

- 📂 **Folder structures** (Clean architecture, modular separation)
- 🐳 **Docker & Docker Compose** with multi-stage production builds
- 🛡️ **Graceful Shutdown & Signal Handling** to avoid abruptly terminated connections
- 📝 **Configuration layers** (Environment variables, `.env` validation)
- 🧪 **Makefiles, Linter configs & Git hooks**

**Umaru CLI ships all of this out-of-the-box.** Every template is architected to be immediately deployable and extensible.

### ⚡ Real-World Scaffolding Benchmarks

Unlike `npx` or script-based generators that query remote registries and download packages over the network, Umaru CLI compiles all 23 architectures directly into a single standalone static binary. 

The following real-world measurements compare project scaffolding time with dependency installs skipped (`--skip-install` / `--disable-git`):

| Tool / Scaffolder | Target Architecture | Scaffolding Time | 100% Offline? | Required Runtime |
|---|---|:---:|:---:|:---:|
| `npx create-vite@latest` (cold) | React + TypeScript | **5,333 ms** (5.33 s) | ❌ No (queries npm) | Node.js + npm (100MB+) |
| `npx create-vite` (warm cache) | React + TypeScript | **3,177 ms** (3.18 s) | ❌ No (queries npm) | Node.js + npm (100MB+) |
| `npx create-next-app@latest` | Next.js (App Router + TS) | **3,144 ms** (3.14 s) | ❌ No (queries npm) | Node.js + npm (100MB+) |
| **⚡ Umaru CLI (`umaru init`)** | **React + TypeScript** | **96 ms** (0.09 s) | **✔ Yes (Embedded FS)** | **None (Zero dependencies)** |
| **⚡ Umaru CLI (`umaru init`)** | **Next.js + Tailwind** | **89 ms** (0.08 s) | **✔ Yes (Embedded FS)** | **None (Zero dependencies)** |
| **⚡ Umaru CLI (`umaru init`)** | **Go + Fiber (Clean Arch)** | **56 ms** (0.05 s) | **✔ Yes (Embedded FS)** | **None (Zero dependencies)** |

> 🚀 **Takeaway:** Umaru CLI scaffolds production-grade projects **33x to 55x faster** than traditional `npm create` tools, works completely offline without network latency, and requires zero language runtimes installed to generate boilerplates across Go, Node, Bun, Python, and Rust.

---

## ✨ Features

- 🏎️ **Instantaneous & Lightweight:** Built in Go with zero external runtime dependencies. Compiles to a single static binary.
- 🔌 **Zero Network Reliance:** All 23 starter boilerplates are compiled directly into the binary via `//go:embed`.
- 🧩 **Interactive Addon Wizard:** Modular feature injection (PostgreSQL, SQLite, JWT Auth, Redis Cache).
- ⚙️ **Persistent User Preferences:** Remember your preferred package manager, author, and licenses via `~/.umarurc.json`.
- 🌐 **Remote Template Scaffolding:** Scaffold directly from any GitHub repo via `--from owner/repo`.
- 🎨 **Modern Terminal DX:** Interactive, accessible prompts powered by [Huh](https://github.com/charmbracelet/huh) and styled result cards with [Lipgloss](https://github.com/charmbracelet/lipgloss).
- 🐚 **Dynamic Shell Autocompletions:** Instant completion for template IDs, database drivers, and flags in Bash, Zsh, Fish, and PowerShell.
- 🔄 **One-Command Upgrades:** Built-in self-updater via `umaru upgrade`.
- 📦 **Universal Package Manager Support:** Choose your preferred JS/TS package manager on the fly (`npm`, `pnpm`, `yarn`, `bun`).
- 🩺 **Environment Diagnostics:** Run `umaru doctor` to inspect installed runtimes, package managers, Docker daemon status, and template ecosystem readiness.
- 🔍 **Architecture Deep Inspection:** Inspect directory trees, default ports, and tech stacks of any template with `umaru info <template>`.
- 🛡️ **Pre-Flight Verification:** Proactively checks system dependencies (`git`, `go`, `cargo`, `pnpm`, etc.) beforehand so generation never fails halfway through.
- 🔍 **Dry-Run Mode:** Simulate and inspect every file that would be generated without writing anything to disk.
- 🔤 **Unicode & Transliteration Engine:** Native slugification for Turkish and accented characters (e.g. `Çalışma Projesi` ➔ `calisma-projesi`) for compliant `package.json`, `go.mod`, and `Cargo.toml`.
- 📜 **Verbose Streaming Logs:** Optional live command streaming to monitor dependency installations in real-time.

---

## 📦 Supported Starters

Umaru CLI includes 23 production-ready architectures organized across 4 categories:

### ⚙️ Backend APIs
| Template ID | Technology Stack | Architecture & Included Features |
|---|---|---|
| `go-fiber` | **Go + Fiber v2** | Layered Clean Architecture (`cmd/`, `internal/`), Docker Multi-Stage, `docker-compose`, Graceful Shutdown, CORS, Makefile. |
| `go-gin` | **Go 1.24 + Gin** | Enterprise Clean Architecture, Gin Recovery & Logger, CORS, Graceful Shutdown, Docker & Compose. |
| `go-echo` | **Go + Echo v4** | Clean Architecture (`cmd/`, `internal/routes`, `handlers`, `config`), Docker Multi-Stage, `docker-compose`, Graceful Shutdown, CORS, Makefile. |
| `bun-elysia` | **Bun + Elysia.js + TS** | Ultra-fast TypeScript API, OpenAPI Swagger (`/docs`), CORS, Docker Multi-Stage, `docker-compose`. |
| `fastify-api` | **Fastify + TypeScript** | High-throughput backend, OpenAPI Swagger UI (`/docs`), Strict TS, Docker Multi-Stage, `docker-compose`. |
| `hono-api` | **Hono + TypeScript** | Ultrafast lightweight TypeScript API (Node.js/Bun adapter), CORS, Logger, Docker & Compose. |
| `hono-cloudflare` | **Hono + Cloudflare Workers + TS** | Ultra-fast edge-native serverless API, zero cold-start, Wrangler CLI, KV & D1 bindings, Geo telemetry. |
| `node-express` | **Node.js + TypeScript** | Modular Express architecture (`controllers/`, `routes/`, `middlewares/`), Helmet, Morgan, CORS, Global Error Handler. |
| `nestjs-api` | **NestJS 10 + TypeScript** | Enterprise modular backend, Swagger OpenAPI (`/api/docs`), ValidationPipe, Docker & Compose, Jest test suite. |
| `python-fastapi` | **FastAPI + Pydantic v2** | Versioned API Router (`/api/v1/`), Pydantic models, Interactive OpenAPI Swagger `/docs`, Docker, CORS. |
| `ai-fastapi-starter` | **FastAPI + OpenAI/Ollama + ChromaDB** | Production AI & LLM Streaming API, Server-Sent Events (SSE), Vector DB, Pydantic v2, Docker Compose. |
| `rust-actix` | **Rust + Actix-Web 4** | Safe, ultra-high throughput backend, Serde JSON serialization, Health check endpoints. |
| `rust-axum` | **Rust + Axum 0.7 + Tokio** | Async Tokio runtime, Tower HTTP middleware, Tracing subscriber, Docker Multi-Stage, Graceful Shutdown. |

### 🌐 Frontend Applications
| Template ID | Technology Stack | Included Features |
|---|---|---|
| `nextjs-tailwind` | **Next.js 14 + Tailwind CSS** | App Router, PostCSS, Lucide Icons, TypeScript, Optimized SEO meta defaults. |
| `astro-tailwind` | **Astro 4 + Tailwind CSS** | Content-driven architecture, Zero-JS by default, Markdown/MDX ready, Lucide Icons. |
| `react-vite-ts` | **React 18 + Vite 5 + TS** | Lightning-fast HMR, Strict TypeScript, Lucide Icons, Tailwind CSS. |
| `svelte-vite-ts` | **Svelte 5 + Vite 5 + TS** | Modern Runes reactivity (`$state`), Tailwind CSS, Lucide Icons, Vite HMR. |
| `vue-vite-ts` | **Vue 3 + Vite 5 + TS** | Composition API (`<script setup>`), Pinia State Management, Tailwind CSS, Lucide Icons. |

### 📦 Fullstack Applications & Monorepos
| Template ID | Technology Stack | Included Features |
|---|---|---|
| `go-htmx` | **Go Fiber + HTMX 2.0 + Tailwind** | Modern Hypermedia Stack, Fiber HTML templates, Zero-JS dynamic state, Tailwind CSS, Docker & Makefile. |
| `fullstack-go-react` | **Go Fiber + React Vite + TS** | Monorepo structure (`apps/api`, `apps/web`), Live API Proxy, Unified Docker Compose, Makefile. |
| `fullstack-ts-monorepo` | **Hono API + React Vite + TS** | High-performance TypeScript Monorepo, Hono backend, React 18 frontend, Tailwind CSS, Docker Compose. |

### ⚡ CLI & Terminal Tools
| Template ID | Technology Stack | Architecture & Included Features |
|---|---|---|
| `go-tui` | **Go + Bubble Tea + Lipgloss** | Full-screen interactive Terminal User Interface, multi-tab navigation, dynamic resizing, Bubbles & Lipgloss design. |
| `go-cli` | **Go + Cobra + Bubble Tea + Lipgloss** | Modern CLI & interactive TUI, Viper configuration, Lipgloss Dracula styles, Makefile & multi-stage build. |

---

## 🧩 Interactive Addon Wizard

When scaffolding backend or fullstack projects, Umaru CLI can automatically inject modular infrastructure:

- 🐘 **Database Driver:** `PostgreSQL` (connection pool & healthcheck) or `SQLite` (embedded WAL mode).
- 🔐 **Authentication:** `JWT` (claim generation & verification middleware).
- 🔴 **Cache:** `Redis` (client connection pool & ping).
- 🐳 **Containerization:** `Docker` (multi-stage `Dockerfile`, `docker-compose.yml`, `.dockerignore`).
- 🤖 **CI/CD Pipeline:** `GitHub Actions` (automated test, lint, and build workflows tailored to Go, Node, Python, and Rust).

```bash
# Non-interactive addon specification
umaru init my-backend -t go-fiber --db postgres --auth jwt --redis

# Skip addon prompts during interactive initialization
umaru init my-backend --no-addons
```

### ➕ Inject Addons into Existing Projects (`umaru add`)

Already have an existing project? Umaru CLI automatically detects your language and framework (Go, Node.js, Bun, Python, Rust) and injects modular addons into your existing codebase. You can even stack multiple addons simultaneously in a single pass:

```bash
# Interactive multi-select addon wizard
umaru add

# Add GitHub Actions automated CI/CD pipeline
umaru add ci

# Add containerization to your project
umaru add docker

# Stack multiple addons at once
umaru add postgres redis jwt docker ci

# Direct single addon injection
umaru add redis
umaru add jwt
umaru add sqlite
umaru add docker
umaru add ci

# Overwrite existing addon files
umaru add docker --force
```

---

## ⚙️ Global Configuration

Save your personal defaults to `~/.umarurc.json` so you never have to re-type them:

```bash
# Set your default package manager (npm, pnpm, yarn, bun)
umaru config set pm pnpm

# Set your default project author
umaru config set author "Baran Igsiz"

# View all saved preferences in a table
umaru config list

# Reset all preferences to defaults
umaru config reset
```

---

## 🩺 System Diagnostics (`umaru doctor`)

Verify your development environment, detect installed runtimes and package managers, inspect Docker status, and see exactly which templates are ready to scaffold:

```bash
# Standard diagnostic check
umaru doctor

# Verbose mode (displays binary file paths and troubleshooting info)
umaru doctor --verbose
```

**What Umaru Doctor inspects:**
- 🛠️ **Version Control:** Git detection & version check.
- ⚡ **Runtimes:** Go, Node.js, Python, Cargo (Rust).
- 📦 **Package Managers:** npm, pnpm, yarn, bun, pip.
- 🐳 **Containers:** Docker CLI, Docker Compose, and live Docker Daemon status.
- 📊 **Template Readiness:** Percentage calculation of ready vs. missing tooling across all starter templates.
- 💡 **Actionable Tips:** Direct installation links and commands for any missing tools.

---

## 🚀 Installation

### ⚡ Quick Install (Recommended - Zero Dependencies)

Install the pre-compiled binary instantly in one command:

**Linux & macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Baranigsiz/UmaruCLI/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Baranigsiz/UmaruCLI/main/install.ps1 | iex
```

### 🍺 Via Homebrew (macOS & Linux)

```bash
# Install directly via tap
brew install Baranigsiz/UmaruCLI/umaru

# Or add the tap first and install
brew tap Baranigsiz/umaru https://github.com/Baranigsiz/UmaruCLI
brew install umaru
```

### 🍨 Via Scoop (Windows)

```powershell
# Install directly via manifest URL
scoop install https://raw.githubusercontent.com/Baranigsiz/UmaruCLI/main/bucket/umaru.json

# Or add as bucket and install
scoop bucket add umaru https://github.com/Baranigsiz/UmaruCLI
scoop install umaru
```

### 1. Via Go Install (Any Platform with Go)
```bash
go install github.com/Baranigsiz/UmaruCLI@latest
```

### 2. Pre-Compiled Binaries (Latest: [v1.9.0](https://github.com/Baranigsiz/UmaruCLI/releases/tag/v1.9.0))
Download pre-built binary archives directly from the [GitHub Releases](https://github.com/Baranigsiz/UmaruCLI/releases):

| Platform | Architecture | Binary Archive | Direct Download |
|---|---|---|---|
| **Windows** | `x86_64` (amd64) | `.zip` (`umaru.exe`) | [umaru_1.9.0_windows_amd64.zip](https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_windows_amd64.zip) |
| **Windows** | `ARM64` | `.zip` (`umaru.exe`) | [umaru_1.9.0_windows_arm64.zip](https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_windows_arm64.zip) |
| **macOS** | Apple Silicon (`arm64`) | `.tar.gz` (`umaru`) | [umaru_1.9.0_darwin_arm64.tar.gz](https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_darwin_arm64.tar.gz) |
| **macOS** | Intel (`x86_64`) | `.tar.gz` (`umaru`) | [umaru_1.9.0_darwin_amd64.tar.gz](https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_darwin_amd64.tar.gz) |
| **Linux** | `x86_64` (amd64) | `.tar.gz` (`umaru`) | [umaru_1.9.0_linux_amd64.tar.gz](https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_linux_amd64.tar.gz) |
| **Linux** | `ARM64` | `.tar.gz` (`umaru`) | [umaru_1.9.0_linux_arm64.tar.gz](https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_linux_arm64.tar.gz) |

### 3. Build from Source
```bash
git clone https://github.com/Baranigsiz/UmaruCLI.git
cd UmaruCLI
go build -o umaru main.go

# Optional: Move to your local bin (Linux/macOS)
mv umaru /usr/local/bin/
```

---

## 💻 Usage

### 🎯 Interactive Mode (Recommended)
Simply run `umaru init` and follow the interactive wizard:
```bash
umaru init
```

```text
? What is your project named? my-awesome-api
? Select a category: ⚙️ Backend APIs
? Choose a starter template: Go Fiber API (Production-Ready)
? Choose a database addon: PostgreSQL (Production-ready relational DB)
? Choose an authentication addon: JWT (JSON Web Token authentication)
? Include Redis cache support? Yes

✔ Scaffolding my-awesome-api using Go Fiber API...
✔ Initializing Git repository...
✔ Installing dependencies...

✨ Project Scaffolding Complete!
  📁 Project:    my-awesome-api
  📦 Template:   Go Fiber API (Production-Ready)
  📍 Directory:  my-awesome-api
  🧩 Addons:     DB: postgres, Auth: jwt, Cache: Redis

Next steps to get started:
  1. cd my-awesome-api
  2. go run cmd/api/main.go
```

---

### 🔍 Template Deep Inspection (`umaru info`)

Inspect any starter template before scaffolding to explore its directory architecture, default ports, install/run commands, and compatible addons:

```bash
# Direct template inspection
umaru info go-fiber
umaru info fullstack-ts-monorepo

# Interactive template selector
umaru info
```

---

### ⚡ Non-Interactive & CI/CD Scripting
Provide arguments to bypass prompts for automated workflows:

```bash
# 1. Initialize a Go Fiber API with Postgres & JWT in the current folder
umaru init . -t go-fiber --db postgres --auth jwt --no-git --skip-install

# 2. Scaffold with automatic initial Git commit
umaru init my-elysia-api -t bun-elysia --commit

# 3. Use command aliases ('umaru new' or 'umaru create')
umaru new my-backend -t go-fiber
umaru create my-frontend -t react-vite-ts

# 4. Scaffold a React + Vite application with Bun package manager
umaru init my-frontend -t react-vite-ts -p bun

# 5. Scaffold directly from a remote GitHub repository
umaru init my-custom-app --from username/my-custom-starter

# 6. Simulate file generation without writing to disk
umaru init test-app -t node-express --dry-run

# 7. Stream live dependency installation output
umaru init payment-service -t nestjs-api -p pnpm -v

# 8. Scaffold complete production stack with Docker & CI/CD out of the box
umaru init my-prod-service -t go-fiber --docker --ci --db postgres --redis
```

---

### 🐚 Shell Autocompletion

Umaru CLI supports dynamic autocompletion for Bash, Zsh, Fish, and PowerShell with single-command automatic profile installation:

```bash
# Automatically detect your shell and install autocompletions to your profile
umaru completion --install

# Or specify your target shell explicitly
umaru completion zsh --install
umaru completion powershell --install
umaru completion bash --install
umaru completion fish --install

# Manual one-time session loading:
# Bash
source <(umaru completion bash)

# Zsh
eval "$(umaru completion zsh)"

# Fish
umaru completion fish | source

# PowerShell
umaru completion powershell | Out-String | Invoke-Expression
```

---

### 🔄 Self-Upgrade

Keep Umaru CLI up-to-date with the latest templates and improvements:

```bash
# Check and upgrade to the latest release automatically
umaru upgrade

# Check if a new version is available without installing
umaru upgrade --check
```

---

### 📋 List Starters & Check Version

```bash
# View all available starter templates in a formatted table
umaru list

# Filter templates by category (Frontend, Backend, Fullstack, CLI)
umaru list -c frontend
umaru list -c backend
umaru list -c fullstack

# Display Umaru CLI version and build metadata
umaru version
```

---

## ⚙️ CLI Flags & Options

| Flag | Short | Default | Description |
|---|:---:|:---:|---|
| `--template` | `-t` | `""` | Specify template ID directly (e.g. `go-fiber`, `react-vite-ts`) |
| `--package-manager` | `-p` | `""` | Package manager for Node.js starters (`npm`, `pnpm`, `yarn`, `bun`) |
| `--from` | | `""` | Scaffold directly from a remote Git repository or GitHub shorthand |
| `--db` | | `""` | Inject database driver addon (`postgres`, `sqlite`, `none`) |
| `--auth` | | `""` | Inject authentication middleware addon (`jwt`, `none`) |
| `--redis` | | `false` | Inject Redis caching client module |
| `--docker` | | `false` | Inject Docker multi-stage build & Docker Compose containerization |
| `--ci` | | `false` | Inject GitHub Actions CI/CD pipeline workflow |
| `--no-addons` | | `false` | Skip interactive addon configuration wizard |
| `--dry-run` | | `false` | Simulate generation and list files without creating them |
| `--verbose` | `-v` | `false` | Stream live installation outputs to stdout/stderr |
| `--no-git` | | `false` | Skip automatic `git init` |
| `--skip-install` | | `false` | Skip automatic package/dependency installation |
| `--force` | `-f` | `false` | Overwrite existing files in non-empty target directory |
| `--help` | `-h` | | Display help and usage information |

---

## 🛠️ Extensibility & Custom Templates

Umaru CLI is designed with the **Open/Closed Principle**. You can add new starters to the engine without modifying any Go code:

### 1. Create a Template Directory
Inside `internal/templates/`, create a new folder (e.g., `internal/templates/my-custom-starter`).

### 2. Add `template.json` Metadata
```json
{
  "name": "My Custom Starter",
  "description": "Production-ready boilerplate for specialized workflows.",
  "category": "Backend",
  "installCommand": ["npm", "install"],
  "runCommand": "npm run dev"
}
```

### 3. Add Boilerplate Files
- Any file ending in `.tmpl` will be parsed by Go's `text/template` engine.
- Available template variables:
  - `{{.ProjectName}}` — Raw project name (e.g., `My Cool App`)
  - `{{.SafeName}}` — Sanitized lowercase slug (e.g., `my-cool-app`)
  - `{{.ModuleName}}` — Safe Go module identifier (e.g., `my-cool-app`)
  - `{{.TargetDir}}` — Filesystem destination directory
  - `{{.Author}}` — Configured project author name
  - `{{.License}}` — Configured project license

### 4. Build
```bash
go build -o umaru main.go
```
The new template will automatically be listed in `umaru list`, the interactive wizard, and shell autocompletions!

---

## 🗺️ Roadmap (100% Complete!)

- [x] 🌐 **Remote Templates:** Scaffold directly from GitHub repositories (`umaru init --from user/repo`).
- [x] 🔄 **Self-Updater:** Built-in `umaru upgrade` command to automatically update to the latest release.
- [x] 🐚 **Shell Completions:** Native autocompletion scripts for Bash, Zsh, Fish, and PowerShell with dynamic flag suggestions.
- [x] 🧩 **Interactive Addon Wizard:** Optional feature injection (PostgreSQL, SQLite, Redis, JWT Auth).
- [x] ⚙️ **Config File Support:** Global `~/.umarurc.json` configuration manager (`umaru config`).
- [x] 🩺 **System Diagnostics:** Built-in `umaru doctor` to verify developer environments, runtimes, versions, and template readiness.
- [x] 📦 **23 Production Starters:** Go (Fiber, Gin, Echo, HTMX, Cobra CLI, Bubble Tea TUI), Bun (Elysia), TypeScript (Hono, Hono Cloudflare Workers, NestJS, Express, Fastify), Python (FastAPI, AI & LLM Streaming), Rust (Actix, Axum), Frontend (React, Vue 3, Svelte 5, Next.js, Astro), Fullstack (Go + HTMX, Go + React, TypeScript Monorepo).

---

## 🧪 Testing

Run the full test suite across all templates, generators, updaters, and configs:

```bash
go test -v ./...
```

---

## 🤝 Contributing

Contributions make the open-source community an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingStarter`)
3. Commit your Changes (`git commit -m 'feat: add AmazingStarter template'`)
4. Push to the Branch (`git push origin feature/AmazingStarter`)
5. Open a Pull Request

---

## 📄 License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for more information.

<div align="center">
  <sub>Built with ❤️ by <a href="https://github.com/Baranigsiz/UmaruCLI">Baran Igsiz</a> and contributors.</sub>
</div>

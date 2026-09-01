<div align="center">

# ⚡ Umaru CLI

**A lightning-fast, production-grade project scaffolding CLI for modern developers.**

Bootstraps clean architecture backends, modern frontend apps, and monorepos in milliseconds — complete with Docker, Graceful Shutdown, OpenAPI, and interactive terminal UI.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![CI Workflow](https://img.shields.io/github/actions/workflow/status/Baranigsiz/UmaruCLI/ci.yml?branch=main&style=for-the-badge&label=CI&logo=githubactions&logoColor=white)](https://github.com/Baranigsiz/UmaruCLI/actions)
[![Release](https://img.shields.io/github/v/release/Baranigsiz/UmaruCLI?style=for-the-badge&logo=semanticrelease&logoColor=white&color=7D56F4)](https://github.com/Baranigsiz/UmaruCLI/releases)
[![License](https://img.shields.io/badge/License-MIT-emerald?style=for-the-badge)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge)](CONTRIBUTING.md)

<br />

```text
  _   _                               ____ _     ___ 
 | | | |_ __ ___   __ _ _ __ _   _   / ___| |   |_ _|
 | | | | '_ ` _ \ / _` | '__| | | | | |   | |    | | 
 | |_| | | | | | | (_| | |  | |_| | | |___| |___ | | 
  \___/|_| |_| |_|\__,_|_|   \__,_|  \____|_____|___|
                                                     
          Production Scaffolding in Milliseconds
```

[✨ Features](#-features) • [📦 Starters](#-supported-starters) • [🧩 Addon Wizard](#-interactive-addon-wizard) • [⚙️ Global Config](#️-global-configuration) • [🚀 Installation](#-installation) • [💻 Usage](#-usage) • [🐚 Autocompletion](#-shell-autocompletion) • [🔄 Self-Upgrade](#-self-upgrade) • [🛠️ Extensibility](#️-extensibility--custom-templates)

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

---

## ✨ Features

- 🏎️ **Instantaneous & Lightweight:** Built in Go with zero external runtime dependencies. Compiles to a single static binary.
- 🔌 **Zero Network Reliance:** All 12 starter boilerplates are compiled directly into the binary via `//go:embed`.
- 🧩 **Interactive Addon Wizard:** Modular feature injection (PostgreSQL, SQLite, JWT Auth, Redis Cache).
- ⚙️ **Persistent User Preferences:** Remember your preferred package manager, author, and licenses via `~/.umarurc.json`.
- 🌐 **Remote Template Scaffolding:** Scaffold directly from any GitHub repo via `--from owner/repo`.
- 🎨 **Modern Terminal DX:** Interactive, accessible prompts powered by [Huh](https://github.com/charmbracelet/huh) and styled result cards with [Lipgloss](https://github.com/charmbracelet/lipgloss).
- 🐚 **Dynamic Shell Autocompletions:** Instant completion for template IDs, database drivers, and flags in Bash, Zsh, Fish, and PowerShell.
- 🔄 **One-Command Upgrades:** Built-in self-updater via `umaru upgrade`.
- 📦 **Universal Package Manager Support:** Choose your preferred JS/TS package manager on the fly (`npm`, `pnpm`, `yarn`, `bun`).
- 🛡️ **Pre-Flight Verification:** Proactively checks system dependencies (`git`, `go`, `cargo`, `pnpm`, etc.) beforehand so generation never fails halfway through.
- 🔍 **Dry-Run Mode:** Simulate and inspect every file that would be generated without writing anything to disk.
- 🔤 **Unicode & Transliteration Engine:** Native slugification for Turkish and accented characters (e.g. `Çalışma Projesi` ➔ `calisma-projesi`) for compliant `package.json`, `go.mod`, and `Cargo.toml`.
- 📜 **Verbose Streaming Logs:** Optional live command streaming to monitor dependency installations in real-time.

---

## 📦 Supported Starters

Umaru CLI includes 12 production-ready architectures organized across 3 categories:

### ⚙️ Backend APIs
| Template ID | Technology Stack | Architecture & Included Features |
|---|---|---|
| `go-fiber` | **Go + Fiber v2** | Layered Clean Architecture (`cmd/`, `internal/`), Docker Multi-Stage, `docker-compose`, Graceful Shutdown, CORS, Makefile. |
| `go-gin` | **Go 1.24 + Gin** | Enterprise Clean Architecture, Gin Recovery & Logger, CORS, Graceful Shutdown, Docker & Compose. |
| `node-express` | **Node.js + TypeScript** | Modular Express architecture (`controllers/`, `routes/`, `middlewares/`), Helmet, Morgan, CORS, Global Error Handler. |
| `nestjs-api` | **NestJS 10 + TypeScript** | Enterprise modular backend, Swagger OpenAPI (`/api/docs`), ValidationPipe, Docker & Compose, Jest test suite. |
| `python-fastapi` | **FastAPI + Pydantic v2** | Versioned API Router (`/api/v1/`), Pydantic models, Interactive OpenAPI Swagger `/docs`, Docker, CORS. |
| `rust-actix` | **Rust + Actix-Web 4** | Safe, ultra-high throughput backend, Serde JSON serialization, Health check endpoints. |
| `rust-axum` | **Rust + Axum 0.7 + Tokio** | Async Tokio runtime, Tower HTTP middleware, Tracing subscriber, Docker Multi-Stage, Graceful Shutdown. |

### 🌐 Frontend Applications
| Template ID | Technology Stack | Included Features |
|---|---|---|
| `nextjs-tailwind` | **Next.js 14 + Tailwind CSS** | App Router, PostCSS, Lucide Icons, TypeScript, Optimized SEO meta defaults. |
| `astro-tailwind` | **Astro 4 + Tailwind CSS** | Content-driven architecture, Zero-JS by default, Markdown/MDX ready, Lucide Icons. |
| `react-vite-ts` | **React 18 + Vite 5 + TS** | Lightning-fast HMR, Strict TypeScript, Lucide Icons, Tailwind CSS. |
| `vue-vite-ts` | **Vue 3 + Vite 5 + TS** | Composition API (`<script setup>`), Pinia State Management, Tailwind CSS, Lucide Icons. |

### 📦 Fullstack Monorepos
| Template ID | Technology Stack | Included Features |
|---|---|---|
| `fullstack-go-react` | **Go Fiber + React Vite + TS** | Monorepo structure (`apps/api`, `apps/web`), Live API Proxy, Unified Docker Compose, Makefile. |

---

## 🧩 Interactive Addon Wizard

When scaffolding backend or fullstack projects, Umaru CLI can automatically inject modular infrastructure:

- 🐘 **Database Driver:** `PostgreSQL` (connection pool & healthcheck) or `SQLite` (embedded WAL mode).
- 🔐 **Authentication:** `JWT` (claim generation & verification middleware).
- 🔴 **Cache:** `Redis` (client connection pool & ping).

```bash
# Non-interactive addon specification
umaru init my-backend -t go-fiber --db postgres --auth jwt --redis

# Skip addon prompts during interactive initialization
umaru init my-backend --no-addons
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

## 🚀 Installation

### 1. Via Go Install (Any Platform with Go)
```bash
go install github.com/Baranigsiz/UmaruCLI@latest
```

### 2. Pre-Compiled Binaries
Download the binary for your operating system from the [Releases Page](https://github.com/Baranigsiz/UmaruCLI/releases):

| Platform | Architecture | Binary Format |
|---|---|---|
| **Windows** | `x86_64` (amd64) / `ARM64` | `.zip` (`umaru.exe`) |
| **macOS** | Apple Silicon (`arm64`) / Intel (`x86_64`) | `.tar.gz` (`umaru`) |
| **Linux** | `x86_64` / `arm64` | `.tar.gz` (`umaru`) |

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

### ⚡ Non-Interactive & CI/CD Scripting
Provide arguments to bypass prompts for automated workflows:

```bash
# 1. Initialize a Go Fiber API with Postgres & JWT in the current folder
umaru init . -t go-fiber --db postgres --auth jwt --no-git --skip-install

# 2. Scaffold a React + Vite application with Bun package manager
umaru init my-frontend -t react-vite-ts -p bun

# 3. Scaffold directly from a remote GitHub repository
umaru init my-custom-app --from username/my-custom-starter

# 4. Simulate file generation without writing to disk
umaru init test-app -t node-express --dry-run

# 5. Stream live dependency installation output
umaru init payment-service -t nestjs-api -p pnpm -v
```

---

### 🐚 Shell Autocompletion

Umaru CLI supports dynamic autocompletion for Bash, Zsh, Fish, and PowerShell:

```bash
# Bash
source <(umaru completion bash)

# Zsh
umaru completion zsh > "${fpath[1]}/_umaru"

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
| `--db` | | `""` | Inject database driver addon (`postgres`, `sqlite`, `mongodb`, `none`) |
| `--auth` | | `""` | Inject authentication middleware addon (`jwt`, `none`) |
| `--redis` | | `false` | Inject Redis caching client module |
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

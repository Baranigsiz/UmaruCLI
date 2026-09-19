# Changelog

All notable changes to **Umaru CLI** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v2.0.3] - 2026-09-19

### 🚀 New Features & Enhancements

- **`umaru clean` (Project Sanitizer & Disk Space Reclaimer)**: Automatically scans and safely purges heavy build artifacts, dependencies, and caches (`node_modules`, `target/`, `dist/`, `.next/`, `__pycache__/`, `coverage/`, `.cache/`, `tmp/`) across projects or monorepos with `--dry-run`, `-f, --force`, `-r, --recursive`, `--all`, and `--json` support. Strictly protects `.git` and source code.
- **`umaru dev` (Universal Zero-Config Development Runner)**: Automatically detects project language (Go, Node/TypeScript, Python, Rust, Docker Monorepo), lockfile (`pnpm`, `bun`, `yarn`, `npm`), framework, and entrypoint, and launches the live dev server without requiring developers to remember framework-specific commands. Aliases: `umaru run`, `umaru start`.
- **`umaru config init` (Interactive Setup Wizard)**: Added interactive terminal configuration setup wizard powered by `charmbracelet/huh` to configure default author, preferred package manager, license, and git-init preferences.
- **Global `--debug` & `--config <path>` Flags**: Added global `--debug` flag for verbose error stack traces, and `--config` flag to override `~/.umarurc.json` with a custom configuration file for CI/CD and automation.
- **`umaru ls` Alias**: Added `ls` alias to `umaru list` matching standard CLI conventions.
- **Rich Command Usage Examples**: Added comprehensive `Example:` sections across all commands (`init`, `dev`, `clean`, `add`, `list`, `info`, `doctor`, `config`, `upgrade`, `completion`).

### 🛠️ Bug Fixes & Stability Improvements

- **Transliteration ASCII 'I' Bug**: Removed ASCII 'I' from Turkish transliteration table so standard English uppercase 'I' (e.g. `MyProject`, `Istanbul`, `ID`) is preserved instead of being lowercased to 'i'.
- **Context Cancellation (Ctrl+C Child Process Leak)**: Propagated `cmd.Context()` into `runScaffoldWorkflow` and child process actions (`git`, `npm install`, etc.) so child processes terminate immediately when user cancels via Ctrl+C.
- **Docker Compose V2 Standard**: Removed deprecated `version: '3.8'` from generated `docker-compose.yml` files, aligning with Docker Compose V2 specifications.
- **`Slugify` Consecutive Dash Normalization**: Added regex collapsing for multiple consecutive hyphens (`my---app` -> `my-app`).
- **`.dockerignore` Protection**: Ensured existing `.dockerignore` files are protected from overwrite, matching `Dockerfile` and `docker-compose.yml` behavior.
- **HTTP Download Timeout**: Increased updater HTTP client timeout from 15s to 120s to prevent timeouts on slower network connections.
- **Doctor Panic Recovery**: Added deferred panic recovery guards to concurrent tool diagnostic goroutines.
- **NPM Builtin vs Script Normalization**: Resolved `GetRunCommand` mapping for `npm start` and `npm test` across bun, pnpm, yarn, and npm.

---

## [v2.0.2] - 2026-09-19

### 🚀 New Features & Enhancements

- **Remote Branch & Tag Support (`--from`)**: Added support for branch/tag ref specifications (`owner/repo#branch`, `owner/repo#v1.0.0`) and non-scheme provider URLs (`github.com/...`, `gitlab.com/...`), enabling scaffolding from specific versions.
- **Git Identity Preservation**: Initial commits now automatically preserve developer's configured Git author name/email and signing settings, falling back to Umaru config `author` or `Umaru CLI`.
- **Automated Open-Source `LICENSE` Generation**: Scaffolding now automatically emits standard `LICENSE` files (MIT, Apache-2.0, BSD-3, ISC, GPL-3, Unlicense) customized with the current year and configured author.
- **Interactive Wizard Search**: Added live keyword search (`🔍 Search templates by keyword...`) to the interactive terminal wizard, enabling instant filtering across all 25 starter templates.
- **`umaru config unset <key>`**: Added ability to reset a single configuration key back to default without clearing the entire user config.
- **`umaru version --json`**: Machine-readable JSON output for CLI version, git commit, build timestamp, OS, and architecture.
- **`umaru add --dry-run` & `--skip-install`**: Preview injected addon files before writing to disk and skip dependency installations in offline/CI environments.
- **GitHub API Rate Limit Guard**: Updater now automatically attaches `GITHUB_TOKEN` / `GH_TOKEN` if present in the environment to increase API limits from 60 to 5,000 req/hr.
- **Package Manager Collision Guard**: `umaru upgrade` now detects if the binary was installed via Homebrew (`brew upgrade umaru`) or Scoop (`scoop update umaru`) to avoid corrupting package manager trees.

### 🛠️ Bug Fixes & Stability Improvements

- **Non-Interactive TTY Hang**: Integrated `isatty` terminal detection across `cmd/init.go`, `prompts.go`, `cmd/add.go`, and `cmd/info.go` to fail fast with actionable errors in non-interactive / CI piped environments instead of freezing on prompts.
- **Addon Audit False Positives**: Resolved collision where both PostgreSQL and SQLite were reported as installed simultaneously for Node and Python projects by adding file content inspection. Also ensured Docker is only reported as installed when `Dockerfile` or `docker-compose.yml` actually exists.
- **Python SQLite Database URL & Dependency**: Fixed Python SQLite addon generation to produce `sqlite+aiosqlite:///./app.db` instead of an async PostgreSQL URL, and inject `aiosqlite>=0.20.0`.
- **Python JWT Secret Name Mismatch**: Standardized `security.py` to prioritize `JWT_SECRET` (matching `.env` and `.env.example`) with fallback to `JWT_SECRET_KEY`.
- **Dockerfile Go Entrypoint Resolution**: Dynamically detects Go entrypoints (`cmd/api/main.go`, `cmd/web/main.go`, root `main.go`) to prevent Docker build failures.
- **Test Config Isolation**: Prevented unit tests from mutating or wiping the developer's real `~/.umarurc.json` by adding `SetTestConfigDir()` and `TestMain` isolation.

## [v2.0.1] - 2026-09-17

### 🛠️ Bug Fixes & Stability Improvements

- **Doctor Monorepo Dependency Reporting**: Fixed an issue where `umaru doctor` calculated fullstack monorepos (`fullstack-go-react`, `fullstack-ts-monorepo`) as unready when Node was installed without `npm/pnpm/yarn/bun`, but omitted `"npm/pnpm"` from the missing tools array.
- **Remote Scaffolding File Descriptor Leak**: Replaced deferred file closing inside the `filepath.WalkDir` copy loop with an isolated `copyFile` function, eliminating OS file descriptor exhaustion (`too many open files`) when generating projects from large remote repositories.
- **Non-Interactive Scaffolding Flow (`--yes` / `-y`)**:
  - Automatically defaults to `userCfg.PackageManager` or `"npm"` for Node-based templates when `--package-manager` is omitted, eliminating blocking TUI prompts in CI/CD environments.
  - Automatically assumes `--force` when target directories are non-empty in non-interactive mode to prevent TTY hangs.
- **Rust Addon Prompt Alignment**: Updated the interactive addon wizard (`prompts.go`) for Rust templates (`rust-axum`, `rust-actix`) to only present supported container and CI workflows, eliminating prompts for unsupported code-level DB and JWT generators.
- **Windows Updater Cleanup**: Added `CleanupOldExecutable()` to automatically purge leftover `umaru.exe.old` binaries on application startup following a self-upgrade on Windows.

## [v2.0.0] - 2026-09-11

### 🌟 Major Highlights & New Features

- **25 Production-Grade Starters**: Expanded the template ecosystem from 20 to 25 enterprise-ready starter architectures across 5 distinct categories.
- **🖥️ Brand-New Desktop Category**: Added cross-platform native desktop application scaffolding with Tauri v2 and Rust.
- **⚡ Edge-Native Serverless**: Added zero cold-start edge computing starter powered by Cloudflare Workers and Hono.
- **🤖 Enterprise AI & RAG Pipeline**: Added complete document chunking, semantic vector search, citations, and streaming SSE responses.
- **🚀 Enhanced Scaffolding DX**:
  - Added `--docker` flag to automatically containerize any backend or monorepo project in a single command.
  - Added `--ci` flag to automatically inject GitHub Actions CI/CD workflows tailored to the project stack.
  - Added `--category` (`-c`) filter flag to `umaru list` with interactive shell autocompletion.
  - Integrated Docker and CI questions into the interactive Charmbracelet Huh terminal wizard.

### 📦 New Starter Templates

1. **`go-htmx` (Fullstack)**:
   - Modern hypermedia stack with Go Fiber v2, HTML template engine, and HTMX 2.0.
   - Live interactive state without heavy JavaScript (reactive counter and todo list partials).
   - Tailwind CSS standalone build, multi-stage `Dockerfile`, `docker-compose.yml`, and `Makefile`.

2. **`go-tui` (CLI)**:
   - Full-screen interactive Terminal User Interface built with Bubble Tea v1.2 and Lipgloss.
   - Dynamic terminal window resizing, multi-tab navigation (`Dashboard`, `Metrics`, `Settings`).
   - Bubbles spinners, progress bars, and Dracula color theme.

3. **`hono-cloudflare` (Backend)**:
   - Edge-native serverless API running on Cloudflare Workers using Hono v4.
   - Wrangler CLI configuration (`wrangler.toml`), TypeScript strict compilation.
   - Pre-configured Cloudflare KV key-value store, Cloudflare D1 SQL database bindings, and Geo-telemetry (`cf.country`, `cf.city`).

4. **`tauri-desktop` (Desktop)**:
   - Ultra-lightweight native desktop app built with Tauri v2, React 18, Tailwind CSS, and Rust.
   - Type-safe Rust IPC bridge (`greet`, `get_system_info` commands).
   - Tauri v2 capability security model (`capabilities/default.json`).
   - Vite HMR development server and production native bundle packaging.

5. **`ai-rag-agent` (Backend)**:
   - Enterprise Retrieval-Augmented Generation (RAG) backend with FastAPI, LangChain, and ChromaDB.
   - Document chunking and ingestion endpoint (`POST /api/v1/rag/ingest`).
   - Semantic similarity vector search with source citations (`POST /api/v1/rag/query`).
   - Real-time Server-Sent Events token streaming (`POST /api/v1/rag/chat-stream`).
   - Flexible dual-mode vector store (in-process embedded SQLite or standalone ChromaDB container).

### 🛠️ Improvements & Bug Fixes

- **Docker Compose Regex Parsing**: Replaced brittle `strings.Split(content, "volumes:")` with multiline anchored regex `(?m)^volumes:` to prevent corruption of nested volume bindings (e.g. ChromaDB or PostgreSQL container volumes).
- **Monorepo CI Workflow Paths**: Fixed GitHub Actions workflow generation in monorepos (`fullstack-go-react`, `fullstack-ts-monorepo`) to always write to root `.github/workflows/ci.yml` instead of subproject folders.
- **Monorepo Detection**: Fixed framework detector (`detector.go`) recognizing fullstack monorepos as plain Express apps.
- **Git Injection Protection**: Hardened `NormalizeGitURL` against options starting with `-` to protect against command injection via remote template URLs.
- **Diagnostics & Doctor**:
  - Added explicit Bun runtime requirement for `bun-elysia`.
  - Added OS-aware Docker guidance on Windows (`Start Docker Desktop`).
  - Added automatic fallback between `pip` and `pip3` on Linux/macOS.

---

## [v1.9.0] - 2026-09-08

### Added
- Interactive addon injector (`umaru add`) supporting Docker, CI/CD, PostgreSQL, SQLite, Redis, and JWT.
- Persistent configuration manager (`umaru config set/get/list`) backed by `~/.umarurc.json`.
- Environment diagnostics command (`umaru doctor`) with concurrent runtime and package manager checks.
- Deep architecture inspector (`umaru info <template>`) displaying ASCII tree hierarchies and network ports.

---

## [v1.0.0] - 2026-08-20

### Initial Release
- Lightning-fast scaffolding for 20 production architectures in Go, TypeScript, Node.js, Python, and Rust.
- Multi-stage Dockerfile and Docker Compose generators.
- Unicode and Turkish transliteration slugification engine.
- Shell auto-completions for Bash, Zsh, PowerShell, and Fish.

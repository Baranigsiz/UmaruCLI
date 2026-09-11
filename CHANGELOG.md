# Changelog

All notable changes to **Umaru CLI** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

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

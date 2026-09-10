# Contributing to Umaru CLI ⚡

Thank you for your interest in contributing to **Umaru CLI**! We welcome contributions from the community to help make Umaru the fastest, most reliable project scaffolding CLI.

---

## 📜 Table of Contents

- [Code of Conduct](#-code-of-conduct)
- [How to Contribute](#-how-to-contribute)
- [Development Setup](#-development-setup)
- [Architecture Overview](#-architecture-overview)
- [Adding a New Starter Template](#-adding-a-new-starter-template)
- [Adding a New Infrastructure Addon](#-adding-a-new-infrastructure-addon)
- [Running Tests & Linting](#-running-tests--linting)
- [Pull Request Guidelines](#-pull-request-guidelines)

---

## 🤝 Code of Conduct

Please be respectful and considerate of other contributors. We strive to maintain an open, welcoming, and inclusive environment.

---

## 💡 How to Contribute

- 🐛 **Report Bugs:** Open an issue with reproduction steps, your OS, Go version, and terminal output.
- 💡 **Suggest Enhancements:** Open an issue describing the feature, why it is needed, and how it improves developer experience.
- 📦 **Add Starter Templates:** Propose and add new production-ready templates (e.g. Bun, Elysia, Go CLI, etc.).
- 🧩 **Add Addons:** Expand modular infrastructure addons (e.g. MySQL, ORM drivers, Dockerfile generators).
- 📝 **Improve Documentation:** Enhance documentation, guides, and comments.

---

## 🛠️ Development Setup

### Prerequisites
- **Go:** `1.24` or higher installed ([go.dev/dl](https://go.dev/dl/))
- **Git:** Installed and available in your `PATH`
- *(Optional)* **Node.js / Python / Docker:** For testing framework-specific templates and addons.

### Clone and Run
```bash
# 1. Clone your fork
git clone https://github.com/<your-username>/UmaruCLI.git
cd UmaruCLI

# 2. Download Go dependencies
go mod download

# 3. Test run the CLI locally
go run main.go --help

# 4. Build a local binary
go build -o umaru main.go
./umaru list
```

---

## 🏗️ Architecture Overview

The codebase is organized into clean, focused packages:

```text
UmaruCLI/
├── main.go                     # Entrypoint (delegates to cmd.Execute())
├── cmd/                        # Cobra CLI command definitions
│   ├── root.go                 # Base root command
│   ├── init.go                 # 'umaru init' (aliases: 'new', 'create')
│   ├── add.go                  # 'umaru add' (modular addon injector)
│   ├── doctor.go               # 'umaru doctor' (environment diagnostics)
│   ├── info.go                 # 'umaru info' (template architecture inspection)
│   ├── list.go                 # 'umaru list' (starter templates catalog)
│   ├── config.go               # 'umaru config' (persistent ~/.umarurc.json)
│   ├── upgrade.go              # 'umaru upgrade' (GitHub self-updater)
│   └── completion.go           # 'umaru completion' (Shell completions)
└── internal/
    ├── templates/              # Embedded template assets via //go:embed
    ├── generator/              # File generation, slugification, addon injection, git cloning
    ├── doctor/                 # Diagnostic runner & Lipgloss report renderer
    ├── prompts/                # Charmbracelet Huh interactive TUI forms
    ├── ui/                     # Terminal cards, ASCII art, Lipgloss styling
    ├── actions/                # Cross-platform execution (git init, npm install, etc.)
    ├── checks/                 # Pre-flight dependency validation
    └── updater/                # GitHub release fetching and binary replacement
```

---

## 📦 Adding a New Starter Template

All starter templates are compiled directly into the binary using Go's `//go:embed all:*`.

To add a new starter template:

1. **Create Template Directory:**
   Add a new folder inside `internal/templates/<template-id>/` (e.g. `internal/templates/bun-elysia/`).

2. **Add `template.json`:**
   Inside your template folder, create a `template.json` file:
   ```json
   {
     "name": "Bun Elysia API (High-Performance)",
     "description": "Blazing fast TypeScript REST API with Bun and Elysia.js",
     "category": "Backend",
     "installCommand": ["bun", "install"],
     "runCommand": "bun run dev"
   }
   ```

3. **Add Boilerplate Files:**
   Add your template files. Any file that requires dynamic variable substitution should end with `.tmpl` (e.g., `package.json.tmpl` or `README.md.tmpl`).
   
   Available template variables:
   - `{{.ProjectName}}`: Human-readable project name (e.g. `My Awesome App`)
   - `{{.SafeName}}`: Lowercase slug (e.g. `my-awesome-app`)
   - `{{.ModuleName}}`: Go module identifier or package name
   - `{{.Author}}`: Author name (from `~/.umarurc.json` or blank)
   - `{{.License}}`: Project license (default: `MIT`)

4. **Register Ports & Addons (Optional):**
   In [internal/templates/info.go](internal/templates/info.go), update:
   - `GetTemplatePorts`: Add the default network port.
   - `GetSupportedAddons`: Declare which addons this template supports.

5. **Verify with Tests:**
   Umaru CLI automatically runs generative tests across all available templates:
   ```bash
   go test -v ./internal/templates
   go test -v ./internal/generator
   ```

---

## 🧩 Adding a New Infrastructure Addon

Addons allow developers to inject modular components (databases, auth, cache) into projects:

1. **Update Addon Configuration:**
   In [internal/generator/addons.go](internal/generator/addons.go), update `AddonConfig` and `TemplateSupportsAddons`.

2. **Implement Generation Logic:**
   In `GenerateAddons` inside `internal/generator/addons.go`, add code generation templates for supported languages (Go, TypeScript/Node, Python).

3. **Register in CLI:**
   Update [cmd/add.go](cmd/add.go) arguments and interactive `huh.NewMultiSelect` prompt.

4. **Add Unit Tests:**
   Add tests to [internal/generator/addons_test.go](internal/generator/addons_test.go).

---

## 🧪 Running Tests & Linting

Before opening a pull request, ensure all tests pass and code is formatted cleanly:

```bash
# Run all unit tests
go test ./...

# Run tests with verbose output and race detector
go test -v -race ./...

# Run static analysis
go vet ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

---

## 🚀 Pull Request Guidelines

1. **Branch Naming:** Use descriptive branch names:
   - `feat/add-bun-elysia-template`
   - `fix/git-commit-quoting`
   - `docs/update-readme`

2. **Commit Messages:** Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat: add bun-elysia starter template`
   - `fix: resolve windows path quoting in git actions`
   - `docs: add contributing guide`
   - `test: add unit tests for cmd package`

3. **Quality Checklist:**
   - [ ] All tests pass (`go test ./...`).
   - [ ] Code is formatted with `gofmt` / `goimports`.
   - [ ] No unnecessary external dependencies added to `go.mod`.
   - [ ] Documentation updated if relevant.

---

Thank you for helping make **Umaru CLI** awesome! ⚡

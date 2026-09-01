# Umaru CLI 🚀

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![CI](https://github.com/Baranigsiz/UmaruCLI/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/Baranigsiz/UmaruCLI/actions/workflows/release.yml/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Umaru** is a lightning-fast, production-grade project scaffolding CLI built with Go. It instantly bootstraps modern web and API projects configured with industry-standard architectures, Docker support, graceful shutdown, environment configurations, and interactive Charmbracelet terminal UI.

---

## ✨ Features

- 🏎️ **Lightning Fast**: High performance, single static binary compiled with Go.
- 🏗️ **Production-Ready Architectures**: Scaffolds full clean architectures (Controllers, Routes, Config, Middlewares, Docker & `docker-compose.yml`, Makefiles) instead of bare-bones Hello-World scripts.
- 🎨 **Modern Terminal DX**: Beautiful interactive prompts with [Huh](https://github.com/charmbracelet/huh) and styled Lipgloss result cards with next-step instructions.
- 📂 **Smart Path & Slug Resolution**: Full support for `.` (current directory), path normalization, and automatic slugification for `package.json`, `go.mod`, and `Cargo.toml`.
- 🛡️ **Pre-Flight Checks**: Checks for required dependencies (`git`, `npm`, `cargo`, `go`) beforehand so generation never fails halfway through.
- 🔌 **Zero External Dependencies**: All starter templates are compiled directly into the binary using Go's `embed.FS`.

---

## 📦 Supported Templates

Umaru CLI natively provides batteries-included starter architectures:

| Template | Stack | Architecture Features |
|---|---|---|
| 🏎️ **Go Fiber API** | Go + Fiber v2 | Clean architecture (`cmd/`, `internal/`), Docker multi-stage build, `docker-compose`, Makefile, Graceful Shutdown, CORS & Logger. |
| 🌐 **Node.js Express** | Node.js + TypeScript | Modular architecture (`controllers/`, `routes/`, `middlewares/`), Docker, Helmet, Morgan, CORS, Error handling. |
| 🐍 **Python FastAPI** | FastAPI + Pydantic v2 | Versioned router (`/api/v1/`), Pydantic models, OpenAPI `/docs`, Docker, `docker-compose`, CORS. |
| ▲ **Next.js** | Next.js 14 + Tailwind CSS | App Router, Tailwind CSS, Lucide icons, PostCSS, TypeScript. |
| ⚛️ **React** | React 18 + Vite 5 + TS | Fast development setup, Lucide icons, strict TypeScript. |
| 🦀 **Rust Actix Web** | Rust + Actix-Web 4 | Safe and ultra-fast backend, Serde JSON serialization, healthchecks. |

---

## 🚀 Installation

### Using Pre-Compiled Binaries (Recommended)
Download the latest binary for Windows, macOS, or Linux directly from the [Releases Page](https://github.com/Baranigsiz/UmaruCLI/releases).

1. Download the archive for your OS.
2. Extract the binary (`umaru.exe` or `umaru`).
3. Add it to your system's `PATH`.

### From Source
```bash
git clone https://github.com/Baranigsiz/UmaruCLI.git
cd UmaruCLI
go build -o umaru main.go
# Move it to your bin folder (macOS/Linux)
mv umaru /usr/local/bin/
```

---

## 💻 Usage

### Interactive Mode
Start a new project interactively with terminal prompts:
```bash
umaru init
```

### Non-Interactive / Scriptable Mode
Pass arguments directly for instant CI/CD or scripted scaffolding:
```bash
# Initialize a production-ready Go Fiber API in the current directory
umaru init . -t go-fiber --no-git --skip-install

# Scaffold a React app and overwrite if directory is not empty
umaru init my-frontend -t react-vite-ts --force
```

### List Available Templates
Explore all built-in starter templates:
```bash
umaru list
```

### Check Version
```bash
umaru version
```

### CLI Flags for `umaru init`
| Flag | Short | Description |
|---|---|---|
| `--template` | `-t` | Specify the starter template ID |
| `--no-git` | | Skip Git repository initialization |
| `--skip-install` | | Skip installing package dependencies |
| `--force` | `-f` | Overwrite existing files in target directory |

---

## 🛠️ Architecture & Extensibility

Umaru CLI follows the **Open/Closed Principle**. You can add new templates without modifying any Go code:

1. Create a new folder inside `internal/templates/`.
2. Add a `template.json` file:
```json
{
  "name": "My Custom Template",
  "description": "Clean production starter.",
  "installCommand": ["npm", "install"],
  "runCommand": "npm run dev"
}
```
3. Add your boilerplate files. Files ending with `.tmpl` automatically render template variables (`{{.ProjectName}}`, `{{.SafeName}}`, `{{.ModuleName}}`).
4. Rebuild the project.

---

## 🧪 Testing

```bash
go test -v ./...
```

---

## 📜 Built With
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Huh](https://github.com/charmbracelet/huh) - Interactive prompts & forms
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions & table rendering
- [GoReleaser](https://goreleaser.com/) - Automated release workflow

---

## License
MIT

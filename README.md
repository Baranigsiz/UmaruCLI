# Umaru CLI 🚀

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![CI](https://github.com/Baranigsiz/UmaruCLI/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/Baranigsiz/UmaruCLI/actions/workflows/release.yml/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

Umaru is a professional, lightning-fast, and highly extensible CLI tool built with Go to bootstrap modern developer projects instantly. It provides an interactive terminal UI, smart dependency checks, and scaffolds your projects using industry-standard architectures.

## ✨ Features

- **Lightning Fast**: Built with Go for maximum performance.
- **Data-Driven Templates**: Easily extensible. Add new templates without writing a single line of Go code.
- **Zero External Dependencies**: All templates are embedded directly into the single binary using Go's `embed.FS`.
- **Pre-Flight Checks**: Fails fast if required environment tools (like `git`, `npm`, `cargo`) are missing before generating any files.
- **Interactive UI**: Beautiful, intuitive terminal prompts using Charmbracelet's `huh`.
- **Automated Setup**: Automatically initializes a git repository and installs project dependencies (via npm, go mod, cargo, pip).

## 📦 Supported Templates

Umaru CLI natively supports the following boilerplate architectures out-of-the-box:

- 🏎️ **Go Fiber API** (High performance Go web framework)
- 🌐 **Node.js Express (TypeScript)** (Modern JS backend)
- ⚛️ **React (Vite + TypeScript)** (Blazing fast frontend)
- ▲ **Next.js (App Router + Tailwind)** (Full-stack React)
- 🐍 **Python FastAPI** (Modern, fast Python backend)
- 🦀 **Rust Actix Web** (Extremely fast and safe backend)

## 🚀 Installation

### Using Pre-Compiled Binaries (Recommended)
You can download the latest pre-compiled binaries for Windows, macOS, and Linux directly from our [Releases Page](https://github.com/Baranigsiz/UmaruCLI/releases).

1. Download the archive for your operating system.
2. Extract the binary (`umaru.exe` or `umaru`).
3. Add it to your system's `PATH`.

### From Source
```bash
git clone https://github.com/Baranigsiz/UmaruCLI.git
cd UmaruCLI
go build -o umaru main.go
# Move it to your bin folder
mv umaru /usr/local/bin/
```

## 💻 Usage

### Interactive Mode
Start a new project interactively with beautiful prompts:
```bash
umaru init
```

### Scriptable / Non-Interactive Mode (CI/CD Friendly)
Pass arguments and flags directly to skip interactive menus:
```bash
# Initialize a Go Fiber project skipping git and install steps
umaru init my-api -t go-fiber --no-git --skip-install

# Scaffold a React app and overwrite target directory if not empty
umaru init my-frontend -t react-vite-ts --force
```

### List Available Templates
Explore all built-in starter templates directly from your terminal:
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
| `--skip-install` | | Skip running the package manager / installing dependencies |
| `--force` | `-f` | Overwrite existing files in target directory |


## 🛠️ Architecture & Extensibility

Umaru CLI follows the **Open/Closed Principle**. You can add new templates without modifying the core Go logic.

To add a new template:
1. Create a new folder inside `internal/templates/`.
2. Add a `template.json` file to configure its behavior:
```json
{
  "name": "My Custom Template",
  "description": "Description goes here.",
  "installCommand": ["npm", "install"],
  "runCommand": "npm run dev"
}
```
3. Add your boilerplate files inside the folder. Files ending with `.tmpl` will have variables (like `{{.ProjectName}}`) parsed and the extension stripped automatically.
4. Rebuild the project. The new template will magically appear in the menu!

## 🧪 Testing

The project includes unit tests for core functionalities, utilizing an isolated in-memory/temp filesystem to ensure generation logic remains pristine.
```bash
go test ./...
```

## 📜 Built With
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Huh](https://github.com/charmbracelet/huh) - Interactive prompts & forms
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions & table rendering
- [GoReleaser](https://goreleaser.com/) - Automated CI/CD pipeline

## License
MIT

# Umaru CLI 🚀

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Release](https://img.shields.io/github/v/release/Baranigsiz/UmaruCLI?color=success)
![Build](https://img.shields.io/github/actions/workflow/status/Baranigsiz/UmaruCLI/release.yml?logo=github)
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

Start a new project interactively. The CLI will guide you through the process:
```bash
umaru init
```

Check the version:
```bash
umaru version
```

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
- [Huh](https://github.com/charmbracelet/huh) - Interactive prompts
- [GoReleaser](https://goreleaser.com/) - Automated CI/CD pipeline

## License
MIT
